package image

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/anchore/stereoscope/internal/bus"
	"github.com/anchore/stereoscope/internal/log"
	"github.com/anchore/stereoscope/pkg/event"
	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sylabs/squashfs"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

const SingularitySquashFSLayer = "application/vnd.sylabs.sif.layer.v1.squashfs"

// Layer represents a single layer within a container image.
type Layer struct {
	// layer is the raw layer metadata and content provider from the GCR lib
	layer v1.Layer
	// indexedContent provides index access to the cached and unzipped layer tar
	indexedContent *file.TarIndex
	// Metadata contains select layer attributes
	Metadata LayerMetadata
	// Tree is a filetree that represents the structure of the layer tar contents ("diff tree")
	Tree filetree.Reader
	// SquashedTree is a filetree that represents the combination of this layers diff tree and all diff trees
	// in lower layers relative to this one.
	SquashedTree filetree.Reader
	// fileCatalog contains all file metadata for all files in all layers (not just this layer)
	fileCatalog           *FileCatalog
	SquashedSearchContext filetree.Searcher
	SearchContext         filetree.Searcher
}

// NewLayer provides a new, unread layer object.
func NewLayer(layer v1.Layer) *Layer {
	return &Layer{
		layer: layer,
	}
}

func (l *Layer) uncompressedTarCache(uncompressedLayersCacheDir string) (string, error) {
	if uncompressedLayersCacheDir == "" {
		return "", fmt.Errorf("no cache directory given")
	}

	tarPath := path.Join(uncompressedLayersCacheDir, l.Metadata.Digest+".tar")

	if _, err := os.Stat(tarPath); !os.IsNotExist(err) {
		return tarPath, nil
	}

	rawReader, err := l.layer.Uncompressed()
	if err != nil {
		return "", err
	}

	fh, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("unable to create layer cache dir=%q : %w", tarPath, err)
	}

	if _, err := io.Copy(fh, rawReader); err != nil {
		return "", fmt.Errorf("unable to populate layer cache dir=%q : %w", tarPath, err)
	}

	return tarPath, nil
}

// Read parses information from the underlying layer tar into this struct. This includes layer metadata, the layer
// file tree, and the layer squash tree.
func (l *Layer) Read(catalog *FileCatalog, imgMetadata Metadata, idx int, uncompressedLayersCacheDir string) error {
	var err error
	tree := filetree.New()
	l.Tree = tree
	l.fileCatalog = catalog
	l.Metadata, err = newLayerMetadata(imgMetadata, l.layer, idx)
	if err != nil {
		return err
	}

	log.Debugf("layer metadata: index=%+v digest=%+v mediaType=%+v",
		l.Metadata.Index,
		l.Metadata.Digest,
		l.Metadata.MediaType)

	monitor := trackReadProgress(l.Metadata)

	switch l.Metadata.MediaType {
	case types.OCILayer,
		types.OCIUncompressedLayer,
		types.OCIRestrictedLayer,
		types.OCIUncompressedRestrictedLayer,
		types.DockerLayer,
		types.DockerForeignLayer,
		types.DockerUncompressedLayer:

		tarFilePath, err := l.uncompressedTarCache(uncompressedLayersCacheDir)
		if err != nil {
			return err
		}

		l.indexedContent, err = file.NewTarIndex(
			tarFilePath,
			layerTarIndexer(tree, l.fileCatalog, &l.Metadata.Size, l, monitor),
		)
		if err != nil {
			return fmt.Errorf("failed to read layer=%q tar : %w", l.Metadata.Digest, err)
		}

	case SingularitySquashFSLayer:
		r, err := l.layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("failed to read layer=%q: %w", l.Metadata.Digest, err)
		}
		// defer r.Close() // TODO: if we close this here, we can't read file contents after we return.

		// Walk the more efficient walk if we're blessed with an io.ReaderAt.
		if ra, ok := r.(io.ReaderAt); ok {
			err = file.WalkSquashFS(ra, squashfsVisitor(tree, l.fileCatalog, &l.Metadata.Size, l, monitor))
		} else {
			err = file.WalkSquashFSFromReader(r, squashfsVisitor(tree, l.fileCatalog, &l.Metadata.Size, l, monitor))
		}
		if err != nil {
			return fmt.Errorf("failed to walk layer=%q: %w", l.Metadata.Digest, err)
		}

	default:
		return fmt.Errorf("unknown layer media type: %+v", l.Metadata.MediaType)
	}

	l.SearchContext = filetree.NewSearchContext(l.Tree, l.fileCatalog.Index)

	monitor.SetCompleted()

	return nil
}

// OpenPath reads the file contents for the given path from the underlying layer blob, relative to the layers "diff tree".
// An error is returned if there is no file at the given path and layer or the read operation cannot continue.
func (l *Layer) OpenPath(path file.Path) (io.ReadCloser, error) {
	return fetchReaderByPath(l.Tree, l.fileCatalog, path)
}

// OpenPathFromSquash reads the file contents for the given path from the underlying layer blob, relative to the layers squashed file tree.
// An error is returned if there is no file at the given path and layer or the read operation cannot continue.
func (l *Layer) OpenPathFromSquash(path file.Path) (io.ReadCloser, error) {
	return fetchReaderByPath(l.SquashedTree, l.fileCatalog, path)
}

// FileContents reads the file contents for the given path from the underlying layer blob, relative to the layers "diff tree".
// An error is returned if there is no file at the given path and layer or the read operation cannot continue.
// Deprecated: use OpenPath() instead.
func (l *Layer) FileContents(path file.Path) (io.ReadCloser, error) {
	return fetchReaderByPath(l.Tree, l.fileCatalog, path)
}

// FileContentsFromSquash reads the file contents for the given path from the underlying layer blob, relative to the layers squashed file tree.
// An error is returned if there is no file at the given path and layer or the read operation cannot continue.
// Deprecated: use OpenPathFromSquash() instead.
func (l *Layer) FileContentsFromSquash(path file.Path) (io.ReadCloser, error) {
	return fetchReaderByPath(l.SquashedTree, l.fileCatalog, path)
}

// FilesByMIMEType returns file references for files that match at least one of the given MIME types relative to each layer tree.
// Deprecated: use SearchContext().SearchByMIMEType() instead.
func (l *Layer) FilesByMIMEType(mimeTypes ...string) ([]file.Reference, error) {
	var refs []file.Reference
	refVias, err := l.SearchContext.SearchByMIMEType(mimeTypes...)
	if err != nil {
		return nil, err
	}
	for _, refVia := range refVias {
		if refVia.HasReference() {
			refs = append(refs, *refVia.Reference)
		}
	}
	return refs, nil
}

// FilesByMIMETypeFromSquash returns file references for files that match at least one of the given MIME types relative to the squashed file tree representation.
// Deprecated: use SquashedSearchContext().SearchByMIMEType() instead.
func (l *Layer) FilesByMIMETypeFromSquash(mimeTypes ...string) ([]file.Reference, error) {
	var refs []file.Reference
	refVias, err := l.SquashedSearchContext.SearchByMIMEType(mimeTypes...)
	if err != nil {
		return nil, err
	}
	for _, refVia := range refVias {
		if refVia.HasReference() {
			refs = append(refs, *refVia.Reference)
		}
	}
	return refs, nil
}

func layerTarIndexer(ft filetree.Writer, fileCatalog *FileCatalog, size *int64, layerRef *Layer, monitor *progress.Manual) file.TarIndexVisitor {
	builder := filetree.NewBuilder(ft, fileCatalog.Index)

	return func(index file.TarIndexEntry) error {
		var err error
		var entry = index.ToTarFileEntry()

		var contents = index.Open()
		defer func() {
			if err := contents.Close(); err != nil {
				log.Warnf("unable to close file while indexing layer: %+v", err)
			}
		}()
		metadata := file.NewMetadata(entry.Header, contents)

		// note: the tar header name is independent of surrounding structure, for example, there may be a tar header entry
		// for /some/path/to/file.txt without any entries to constituent paths (/some, /some/path, /some/path/to ).
		// This is ok, and the FileTree will account for this by automatically adding directories for non-existing
		// constituent paths. If later there happens to be a tar header entry for an already added constituent path
		// the FileNode will be updated with the new file.Reference. If there is no tar header entry for constituent
		// paths the FileTree is still structurally consistent (all paths can be iterated even though there may not have
		// been a tar header entry for part of the given path).
		//
		// In summary: the set of all FileTrees can have NON-leaf nodes that don't exist in the FileCatalog, but
		// the FileCatalog should NEVER have entries that don't appear in one (or more) FileTree(s).
		ref, err := builder.Add(metadata)
		if err != nil {
			return err
		}

		if size != nil {
			*(size) += metadata.Size
		}
		fileCatalog.addImageReferences(ref.ID(), layerRef, index.Open)

		if monitor != nil {
			monitor.Increment()
		}
		return nil
	}
}

func squashfsVisitor(ft filetree.Writer, fileCatalog *FileCatalog, size *int64, layerRef *Layer, monitor *progress.Manual) file.SquashFSVisitor {
	builder := filetree.NewBuilder(ft, fileCatalog.Index)

	return func(fsys fs.FS, path string, d fs.DirEntry) error {
		ff, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer ff.Close()

		f, ok := ff.(*squashfs.File)
		if !ok {
			return errors.New("unexpected file type from squashfs")
		}

		metadata, err := file.NewMetadataFromSquashFSFile(path, f)
		if err != nil {
			return err
		}

		fileReference, err := builder.Add(metadata)
		if err != nil {
			return err
		}

		if size != nil {
			*(size) += metadata.Size
		}
		fileCatalog.addImageReferences(fileReference.ID(), layerRef, func() io.ReadCloser {
			r, err := fsys.Open(path)
			if err != nil {
				// The file.Opener interface doesn't give us a way to return an error, and callers
				// don't seem to handle a nil return. So, return a zero-byte reader.
				log.Debug(err)
				return io.NopCloser(bytes.NewReader(nil)) // TODO
			}
			return r
		})

		monitor.Increment()
		return nil
	}
}

func trackReadProgress(metadata LayerMetadata) *progress.Manual {
	p := &progress.Manual{}

	bus.Publish(partybus.Event{
		Type:   event.ReadLayer,
		Source: metadata,
		Value:  progress.Monitorable(p),
	})

	return p
}
