package fileutil

type Copier interface {
	FilteredMultiCopyToTemp(dirs []DirToCopy, filters []string) (string, error)
	FilteredCopyToTemp(dir string, filters []string) (tempDir string, err error)
	CleanUp(tempDir string)
}
