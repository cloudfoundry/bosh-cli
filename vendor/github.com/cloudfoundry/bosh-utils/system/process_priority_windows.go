//go:build windows

// Inspired by github.com/hekmon/processpriority (MIT, Copyright 2024 Edouard Hur).
// Reimplemented inline to avoid the external dependency.

package system

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

const (
	// Windows process priority classes.
	// https://learn.microsoft.com/en-us/windows/win32/procthread/scheduling-priorities
	winPriorityBelowNormal = 0x4000 // BELOW_NORMAL_PRIORITY_CLASS
)

// getProcessPriority returns the priority class of the process with the given pid.
func getProcessPriority(pid int) (int, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return 0, fmt.Errorf("failed to open process: %w", err)
	}
	defer windows.CloseHandle(handle) //nolint:errcheck

	priority, err := windows.GetPriorityClass(handle)
	if err != nil {
		return 0, fmt.Errorf("failed to get priority class: %w", err)
	}
	return int(priority), nil
}

// setProcessPriority sets the priority class of the process with the given pid.
func setProcessPriority(pid int, priority int) error {
	handle, err := windows.OpenProcess(windows.PROCESS_SET_INFORMATION, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("failed to open process: %w", err)
	}
	defer windows.CloseHandle(handle) //nolint:errcheck

	if err = windows.SetPriorityClass(handle, uint32(priority)); err != nil {
		return fmt.Errorf("failed to set priority class: %w", err)
	}
	return nil
}

// lowerProcessPriority sets the child process priority class to BelowNormal.
func (r execCmdRunner) lowerProcessPriority(logTag string, processPid int) error {
	parentPid := os.Getpid()

	parentPriority, err := getProcessPriority(parentPid)
	if err != nil {
		r.logger.Error(logTag, "Error getting priority of the current process (pid %d): %s", parentPid, err)
		return err
	}
	r.logger.Debug(logTag, "Current process priority class is %d", parentPriority)

	r.logger.Debug(logTag, "Setting child process (pid %d) priority to BelowNormal", processPid)
	if err = setProcessPriority(processPid, winPriorityBelowNormal); err != nil {
		r.logger.Error(logTag, "Error setting priority on child process (pid %d): %s", processPid, err)
	}
	return err
}
