//go:build windows

package exec

import (
	"syscall"
	"unsafe"
)

func buildCommandLine(pwshPath, command string) string {
	return syscall.EscapeArg(pwshPath) + " -NoProfile -NonInteractive -Command " + syscall.EscapeArg(command)
}

func inheritableSecurityAttributes() *syscall.SecurityAttributes {
	sa := &syscall.SecurityAttributes{InheritHandle: 1}
	sa.Length = uint32(unsafe.Sizeof(*sa))
	return sa
}

// openNulReadHandle opens the NUL device for use as a child process's
// stdin, so the child never blocks on or observes the parent's console
// input.
func openNulReadHandle() (syscall.Handle, error) {
	path, err := syscall.UTF16PtrFromString("NUL")
	if err != nil {
		return 0, err
	}
	return syscall.CreateFile(path, syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, inheritableSecurityAttributes(),
		syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
}

// createOutputPipe returns a pipe whose write end is inheritable (for
// the child) and whose read end is not (the parent keeps and drains it).
func createOutputPipe() (readEnd, writeEnd syscall.Handle, err error) {
	if err = syscall.CreatePipe(&readEnd, &writeEnd, inheritableSecurityAttributes(), 0); err != nil {
		return 0, 0, err
	}
	if err = syscall.SetHandleInformation(readEnd, syscall.HANDLE_FLAG_INHERIT, 0); err != nil {
		syscall.CloseHandle(readEnd)
		syscall.CloseHandle(writeEnd)
		return 0, 0, err
	}
	return readEnd, writeEnd, nil
}

// startSuspendedProcess launches pwsh with its main thread suspended.
// The caller must AssignProcessToJobObject before ever resuming it.
func startSuspendedProcess(workDir, cmdLine string, stdin, stdoutWrite, stderrWrite syscall.Handle) (syscall.ProcessInformation, error) {
	cmdLinePtr, err := syscall.UTF16PtrFromString(cmdLine)
	if err != nil {
		return syscall.ProcessInformation{}, err
	}
	var workDirPtr *uint16
	if workDir != "" {
		workDirPtr, err = syscall.UTF16PtrFromString(workDir)
		if err != nil {
			return syscall.ProcessInformation{}, err
		}
	}

	si := &syscall.StartupInfo{
		Flags:     syscall.STARTF_USESTDHANDLES,
		StdInput:  stdin,
		StdOutput: stdoutWrite,
		StdErr:    stderrWrite,
	}
	si.Cb = uint32(unsafe.Sizeof(*si))

	var pi syscall.ProcessInformation
	err = syscall.CreateProcess(
		nil, // lpApplicationName: rely on the quoted path as argv[0] in cmdLine instead
		cmdLinePtr,
		nil,  // process security attributes
		nil,  // thread security attributes
		true, // inherit handles
		createSuspended|syscall.CREATE_NEW_PROCESS_GROUP,
		nil, // environment: inherit parent's
		workDirPtr,
		si,
		&pi,
	)
	if err != nil {
		return syscall.ProcessInformation{}, err
	}
	return pi, nil
}
