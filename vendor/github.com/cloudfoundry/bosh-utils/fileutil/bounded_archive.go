package fileutil

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func archiveMemberName(name string) (string, error) {
	if len(name) > 4096 || strings.ContainsAny(name, "\\:\x00") || path.IsAbs(name) {
		return "", fmt.Errorf("invalid archive member path")
	}
	name = path.Clean(name)
	if name == ".." || strings.HasPrefix(name, "../") {
		return "", fmt.Errorf("archive member escapes destination")
	}
	return name, nil
}

func (c tarballCompressor) decompressBounded(source, destination string, options CompressorOptions) (result error) {
	if options.MaxExpandedBytes <= 0 || options.MaxArchiveMembers <= 0 {
		return fmt.Errorf("archive limits must be positive")
	}
	if options.SameOwner || options.StripComponents != 0 || options.PathInArchive != "" {
		return fmt.Errorf("bounded extraction does not support owner or path-selection options")
	}
	if int64(options.MaxArchiveMembers) > (math.MaxInt64-1024*1024)/2048 {
		return fmt.Errorf("archive member limit is too large")
	}
	overhead := int64(options.MaxArchiveMembers)*2048 + 1024*1024
	if overhead < 0 || options.MaxExpandedBytes > math.MaxInt64-overhead-1 {
		return fmt.Errorf("archive limits are too large")
	}
	archive, err := c.fs.OpenFile(source, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer archive.Close()
	input := bufio.NewReader(archive)
	decoded, finish, err := archiveStream(input)
	if err != nil {
		return err
	}
	finished := false
	defer func() {
		if !finished {
			_ = finish(true)
		}
	}()
	// Also bound metadata, padding and trailing compressed data, which do not
	// contribute file payload sizes and must not bypass the byte budget.
	stream := &io.LimitedReader{R: decoded, N: options.MaxExpandedBytes + overhead + 1}
	budget := options.ArchiveBudget
	if budget == nil {
		budget, err = NewArchiveBudget(options.MaxExpandedBytes, options.MaxArchiveMembers, options.MaxManifestBytes)
		if err != nil {
			return err
		}
	}
	boundedStream := budgetReader{reader: stream, budget: budget}
	reader := tar.NewReader(boundedStream)
	root, err := os.OpenRoot(destination)
	if err != nil {
		return err
	}
	defer root.Close()
	var total int64
	count := 0
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if count >= options.MaxArchiveMembers {
			return fmt.Errorf("archive exceeds %d members", options.MaxArchiveMembers)
		}
		count++
		if header.Size < 0 || header.Size > options.MaxExpandedBytes-total {
			return fmt.Errorf("archive exceeds %d expanded bytes", options.MaxExpandedBytes)
		}
		if err := budget.reserve(header.Size); err != nil {
			return err
		}
		total += header.Size
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		name, err := archiveMemberName(header.Name)
		if err != nil {
			return err
		}
		manifestName := strings.TrimRight(strings.ToLower(name), " .")
		if options.MaxManifestBytes > 0 && (manifestName == "release.mf" || manifestName == "job.mf") && header.Size > options.MaxManifestBytes {
			return fmt.Errorf("release manifest exceeds metadata byte limit")
		}
		if name == "." {
			if header.Typeflag == tar.TypeDir {
				continue
			}
			return fmt.Errorf("invalid root archive member")
		}
		if err := root.MkdirAll(path.Dir(name), 0755); err != nil {
			return err
		}
		mode := os.FileMode(header.Mode) & 0777
		switch header.Typeflag {
		case tar.TypeDir:
			if err := root.MkdirAll(name, mode|0700); err != nil {
				return err
			}
			// Working directories remain writable so any later failure can be cleaned up.
			if err := root.Chmod(name, mode|0700); err != nil {
				return err
			}
		// archive/tar reports legacy regular files as TypeReg and presents sparse
		// files with their holes expanded, so both are written as regular files.
		case tar.TypeReg, tar.TypeGNUSparse:
			file, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			_, copyErr := io.CopyN(file, reader, header.Size)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			if _, err := archiveMemberName(path.Join(path.Dir(name), header.Linkname)); err != nil {
				return err
			}
			if path.IsAbs(header.Linkname) || strings.ContainsAny(header.Linkname, "\\:\x00") {
				return fmt.Errorf("invalid archive symlink")
			}
			if err := root.Symlink(header.Linkname, name); err != nil {
				return err
			}
		case tar.TypeLink:
			target, err := archiveMemberName(header.Linkname)
			if err != nil {
				return err
			}
			if err := root.Link(target, name); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported archive member type %d", header.Typeflag)
		}
	}
	if _, err := io.Copy(io.Discard, boundedStream); err != nil {
		return err
	}
	if stream.N == 0 {
		return fmt.Errorf("archive expanded stream exceeds byte/metadata budget")
	}
	// Check the final targets too: manifest symlinks or hardlink aliases must
	// not redirect the subsequent YAML reader to a large ordinary member.
	if options.MaxManifestBytes > 0 {
		for _, name := range []string{"release.MF", "job.MF"} {
			info, err := root.Stat(name)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() || info.Size() > options.MaxManifestBytes {
				return fmt.Errorf("release manifest exceeds metadata byte limit")
			}
			if err := budget.reserveMetadata(info.Size()); err != nil {
				return err
			}
		}
	}
	finished = true
	if err := finish(false); err != nil {
		return err
	}

	return nil
}

func archiveStream(input *bufio.Reader) (io.Reader, func(bool) error, error) {
	signature, _ := input.Peek(6)
	noFinish := func(bool) error { return nil }
	if bytes.HasPrefix(signature, gzipMagic) {
		reader, err := gzip.NewReader(input)
		if err != nil {
			return nil, nil, err
		}
		return reader, func(bool) error { return reader.Close() }, nil
	}
	if bytes.HasPrefix(signature, bzip2Magic) {
		return bzip2.NewReader(input), noFinish, nil
	}
	var command *exec.Cmd
	if bytes.HasPrefix(signature, xzMagic) {
		command = exec.Command("xz", "--decompress", "--stdout", "--memlimit-decompress=256MiB")
	}
	if bytes.HasPrefix(signature, zstdMagic) {
		command = exec.Command("zstd", "--decompress", "--stdout", "--memory=256MB")
	}
	if command == nil {
		return input, noFinish, nil
	}
	command.Stdin = input
	command.WaitDelay = time.Second
	output, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := command.Start(); err != nil {
		return nil, nil, err
	}
	return output, func(abort bool) error {
		if abort {
			_ = command.Process.Kill()
		}
		_ = output.Close()
		return command.Wait()
	}, nil
}
