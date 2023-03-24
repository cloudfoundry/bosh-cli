package rustaudit

import (
	"bytes"
	"compress/zlib"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// This struct is embedded in dependencies produced with rust-audit:
// https://github.com/Shnatsel/rust-audit/blob/bc805a8fdd1492494179bd01a598a26ec22d44fe/auditable-serde/src/lib.rs#L89
type VersionInfo struct {
	Packages []Package `json:"packages"`
}

type DependencyKind string

const (
	Build   DependencyKind = "build"
	Runtime DependencyKind = "runtime"
)

type Package struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Source       string         `json:"source"`
	Kind         DependencyKind `json:"kind"`
	Dependencies []uint         `json:"dependencies"`
	Features     []string       `json:"features"`
}

// Default the Kind to Runtime during unmarshalling
func (p *Package) UnmarshalJSON(text []byte) error {
	type pkgty Package
	pkg := pkgty{
		Kind: Runtime,
	}
	if err := json.Unmarshal(text, &pkg); err != nil {
		return err
	}
	*p = Package(pkg)
	return nil
}

var (
	// Returned if an executable is not a known format
	ErrUnknownFileFormat = errors.New("unknown file format")
	// errNoRustDepInfo is returned when an executable file doesn't contain Rust dependency information
	ErrNoRustDepInfo = errors.New("rust dependency information not found")

	// Headers for different binary types
	elfHeader   = []byte("\x7FELF")
	peHeader    = []byte("MZ")
	machoHeader = []byte("\xFE\xED\xFA")
)

func GetDependencyInfo(r io.ReaderAt) (VersionInfo, error) {
	// Read file header
	header := make([]byte, 16)
	n, err := r.ReadAt(header, 0)
	if n < len(header) || err != nil {
		return VersionInfo{}, ErrUnknownFileFormat
	}

	var x exe
	switch {
	case bytes.HasPrefix(header, elfHeader):
		f, err := elf.NewFile(r)
		if err != nil {
			return VersionInfo{}, ErrUnknownFileFormat
		}
		x = &elfExe{f}
	case bytes.HasPrefix(header, peHeader):
		f, err := pe.NewFile(r)
		if err != nil {
			return VersionInfo{}, ErrUnknownFileFormat
		}
		x = &peExe{f}
	case bytes.HasPrefix(header, machoHeader):
		f, err := macho.NewFile(r)
		if err != nil {
			return VersionInfo{}, ErrUnknownFileFormat
		}
		x = &machoExe{f}
	default:
		return VersionInfo{}, ErrUnknownFileFormat
	}

	data, err := x.ReadRustDepSection()
	if err != nil {
		return VersionInfo{}, err
	}

	// The json is compressed using zlib, so decompress it
	b := bytes.NewReader(data)
	reader, err := zlib.NewReader(b)

	if err != nil {
		return VersionInfo{}, fmt.Errorf("section not compressed: %w", err)
	}

	buf, err := io.ReadAll(reader)
	reader.Close()

	if err != nil {
		return VersionInfo{}, fmt.Errorf("failed to decompress JSON: %w", err)
	}

	var versionInfo VersionInfo
	err = json.Unmarshal(buf, &versionInfo)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("failed to unmarshall JSON: %w", err)
	}

	return versionInfo, nil
}

// Interface for binaries that may have a Rust dependencies section
type exe interface {
	ReadRustDepSection() ([]byte, error)
}

type elfExe struct {
	f *elf.File
}

func (x *elfExe) ReadRustDepSection() ([]byte, error) {
	// Try .dep-v0 first, falling back to .rust-deps-v0 as used in
	// in rust-audit 0.1.0
	depInfo := x.f.Section(".dep-v0")

	if depInfo != nil {
		return depInfo.Data()
	}

	depInfo = x.f.Section(".rust-deps-v0")

	if depInfo == nil {
		return nil, ErrNoRustDepInfo
	}

	return depInfo.Data()
}

type peExe struct {
	f *pe.File
}

func (x *peExe) ReadRustDepSection() ([]byte, error) {
	// Try .dep-v0 first, falling back to rdep-v0 as used in
	// in rust-audit 0.1.0
	depInfo := x.f.Section(".dep-v0")

	if depInfo != nil {
		return depInfo.Data()
	}

	depInfo = x.f.Section("rdep-v0")

	if depInfo == nil {
		return nil, ErrNoRustDepInfo
	}

	return depInfo.Data()
}

type machoExe struct {
	f *macho.File
}

func (x *machoExe) ReadRustDepSection() ([]byte, error) {
	// Try .dep-v0 first, falling back to rust-deps-v0 as used in
	// in rust-audit 0.1.0
	depInfo := x.f.Section(".dep-v0")

	if depInfo != nil {
		return depInfo.Data()
	}

	depInfo = x.f.Section("rust-deps-v0")

	if depInfo == nil {
		return nil, ErrNoRustDepInfo
	}

	return depInfo.Data()
}
