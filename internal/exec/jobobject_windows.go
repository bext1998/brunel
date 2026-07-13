//go:build windows

package exec

import (
	"syscall"
	"unsafe"
)

// Win32 constants not defined in the standard syscall package.
const (
	createSuspended = 0x00000004

	jobObjectLimitActiveProcess  = 0x00000008
	jobObjectLimitProcessMemory  = 0x00000100
	jobObjectLimitKillOnJobClose = 0x00002000

	jobObjectExtendedLimitInformation = 9
)

// jobObjectBasicLimitInformation mirrors JOBOBJECT_BASIC_LIMIT_INFORMATION.
// Field order/types match the Win32 header exactly so Go's amd64 struct
// layout produces identical padding to the C ABI; see the sizeof
// assertions in jobobject_windows_test.go for a cheap regression guard.
type jobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

// ioCounters mirrors IO_COUNTERS.
type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

// jobObjectExtendedLimitInformationT mirrors JOBOBJECT_EXTENDED_LIMIT_INFORMATION.
type jobObjectExtendedLimitInformationT struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

var (
	kernel32Exec = syscall.NewLazyDLL("kernel32.dll")

	procCreateJobObjectW         = kernel32Exec.NewProc("CreateJobObjectW")
	procSetInformationJobObject  = kernel32Exec.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject = kernel32Exec.NewProc("AssignProcessToJobObject")
	procResumeThread             = kernel32Exec.NewProc("ResumeThread")
	procTerminateJobObject       = kernel32Exec.NewProc("TerminateJobObject")
	procIsProcessInJob           = kernel32Exec.NewProc("IsProcessInJob")
)

func createJobObject() (syscall.Handle, error) {
	h, _, callErr := procCreateJobObjectW.Call(0, 0)
	if h == 0 {
		return 0, callErr
	}
	return syscall.Handle(h), nil
}

func setJobLimits(job syscall.Handle, maxProcesses uint32, maxMemoryBytes uint64) error {
	limits := jobObjectExtendedLimitInformationT{
		BasicLimitInformation: jobObjectBasicLimitInformation{
			LimitFlags:         jobObjectLimitKillOnJobClose | jobObjectLimitActiveProcess | jobObjectLimitProcessMemory,
			ActiveProcessLimit: maxProcesses,
		},
		ProcessMemoryLimit: uintptr(maxMemoryBytes),
	}
	ret, _, callErr := procSetInformationJobObject.Call(
		uintptr(job),
		jobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limits)),
		unsafe.Sizeof(limits),
	)
	if ret == 0 {
		return callErr
	}
	return nil
}

func assignProcessToJobObject(job, process syscall.Handle) error {
	ret, _, callErr := procAssignProcessToJobObject.Call(uintptr(job), uintptr(process))
	if ret == 0 {
		return callErr
	}
	return nil
}

// resumeThread returns the thread's previous suspend count on success.
// Unlike most of the Win32 calls in this package, 0 is a valid success
// value here — only 0xFFFFFFFF signals failure.
func resumeThread(thread syscall.Handle) (uint32, error) {
	ret, _, callErr := procResumeThread.Call(uintptr(thread))
	if ret == 0xFFFFFFFF {
		return 0, callErr
	}
	return uint32(ret), nil
}

func terminateJobObject(job syscall.Handle, exitCode uint32) error {
	ret, _, callErr := procTerminateJobObject.Call(uintptr(job), uintptr(exitCode))
	if ret == 0 {
		return callErr
	}
	return nil
}

// isProcessInJob reports whether process is a member of job. The result
// comes back through the out-parameter, not the return value.
func isProcessInJob(process, job syscall.Handle) (bool, error) {
	var result uint32
	ret, _, callErr := procIsProcessInJob.Call(uintptr(process), uintptr(job), uintptr(unsafe.Pointer(&result)))
	if ret == 0 {
		return false, callErr
	}
	return result != 0, nil
}
