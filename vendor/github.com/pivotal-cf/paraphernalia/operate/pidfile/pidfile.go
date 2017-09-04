// Package pidfile provides an ifrit.Runner which will manage your PID file
// lifecycle.
package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/tedsuo/ifrit"
)

type runner struct {
	filename string
}

// NewRunner creates a new runner which will handle your PID file lifecycle.
//
// On startup it will write the current PID to a file and on shutdown it will
// delete the file. If there is already a PID file at the specified path then
// it will check to see if the PID is still valid (still running). If it is
// valid then an error will be returned. If it is not valid then it will be
// cleared and a new PID file will be written. This prevents multiple versions
// of the same server starting.
//
// Locking is in place to make sure this is process-safe.
func NewRunner(filename string) ifrit.Runner {
	return &runner{
		filename: filename,
	}
}

func (r *runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	// acquire locked pidfile to prevent other processes from trying to start
	// with the same pidfile (e.g. monit continuously trying to start)
	pidfile, err := acquirePidfile(r.filename)
	if err != nil {
		return err
	}

	// check for an existing pid
	err = checkForExistingPid(pidfile)
	if err != nil {
		return err
	}

	// write the current pid
	err = writePid(pidfile)
	if err != nil {
		return err
	}

	close(ready)

	<-signals

	return releasePidfile(pidfile)
}

func acquirePidfile(path string) (*os.File, error) {
	// ensure parent dir exists
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	pidfile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(pidfile.Fd()), syscall.LOCK_NB|syscall.LOCK_EX)
	if err != nil {
		return nil, err
	}

	return pidfile, nil
}

func checkForExistingPid(pidfile *os.File) error {
	var existingPid int
	_, err := fmt.Fscanf(pidfile, "%d", &existingPid)
	if err != nil {
		return nil
	}

	process, err := os.FindProcess(existingPid)
	if err != nil {
		return nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return nil
	}

	process.Release()

	return processExistsError{
		Filename: pidfile.Name(),
		Pid:      existingPid,
	}
}

func writePid(pidfile *os.File) error {
	err := pidfile.Truncate(0)
	if err != nil {
		return err
	}

	_, err = pidfile.WriteAt([]byte(fmt.Sprintf("%d", os.Getpid())), 0)
	if err != nil {
		return err
	}

	return nil
}

func releasePidfile(pidfile *os.File) error {
	// remove file while locked
	err := os.Remove(pidfile.Name())
	if err != nil {
		return err
	}

	// release flock
	return pidfile.Close()
}

type processExistsError struct {
	Filename string
	Pid      int
}

func (err processExistsError) Error() string {
	return fmt.Sprintf("pidfile '%s' contains active pid: %d", err.Filename, err.Pid)
}
