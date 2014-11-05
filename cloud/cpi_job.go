package cloud

import (
	"path/filepath"
)

type CPIJob struct {
	JobPath     string
	JobsDir     string
	PackagesDir string
}

func (j CPIJob) ExecutablePath() string {
	return filepath.Join(j.JobPath, "bin", "cpi")
}
