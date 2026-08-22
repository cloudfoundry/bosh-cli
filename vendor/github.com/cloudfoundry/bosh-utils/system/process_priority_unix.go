//go:build !windows

package system

// Inspired by github.com/hekmon/processpriority (MIT, Copyright 2024 Edouard Hur).
// Reimplemented inline to avoid the external dependency.

import (
	"os"
	"runtime"
	"syscall"
)

// setProcessPriority sets the nice value of the process with the given pid.
func setProcessPriority(pid int, nice int) error {
	return syscall.Setpriority(syscall.PRIO_PROCESS, pid, nice)
}

// getProcessPriority returns the nice value of the process with the given pid.
func getProcessPriority(pid int) (int, error) {
	if runtime.GOOS == "linux" {
		knice, err := syscall.Getpriority(syscall.PRIO_PROCESS, pid)
		if err != nil {
			return 0, err
		}

		// Linux: convert syscall.Getpriority()'s "kernel nice" to "user nice"
		//	=> unice = 20 - knice
		// See https://linux.die.net/man/2/getpriority
		return (knice - 20) * -1, nil
	}

	return syscall.Getpriority(syscall.PRIO_PROCESS, pid)
}

// lowerProcessPriority sets the child process nice value to parent + 5, clamped at 19.
func (r execCmdRunner) lowerProcessPriority(logTag string, processPid int) error {
	parentPid := os.Getpid()

	parentNice, err := getProcessPriority(parentPid)
	if err != nil {
		r.logger.Error(logTag, "Error getting priority of the current process (pid %d): %s", parentPid, err)
		return err
	}
	r.logger.Debug(logTag, "Current process nice value is %d", parentNice)

	childNice := min(parentNice+5, 19)
	r.logger.Debug(logTag, "Setting child process (pid %d) nice value to %d", processPid, childNice)

	if err = setProcessPriority(processPid, childNice); err != nil {
		r.logger.Error(logTag, "Error setting priority on child process (pid %d): %s", processPid, err)
	}
	return err
}
