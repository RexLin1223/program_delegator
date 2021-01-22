package windows

import (
	"errors"
	"golang.org/x/sys/windows"
	"scp_delegator/logger"
	"strings"
	"syscall"
	"unsafe"
)

const (
	TH32CS_SNAPPROCESS = 0x00000002
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
)

type WindowsProcess struct {
	ProcessID       int
	ParentProcessID int
	Exe             string
}

func processes() ([]WindowsProcess, error) {
	handle, err := windows.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	// get the first process
	err = windows.Process32First(handle, &entry)
	if err != nil {
		return nil, err
	}

	results := make([]WindowsProcess, 0, 100)
	for {
		results = append(results, newWindowsProcess(&entry))

		err = windows.Process32Next(handle, &entry)
		if err != nil {
			// windows sends ERROR_NO_MORE_FILES on last process
			if err == syscall.ERROR_NO_MORE_FILES {
				return results, nil
			}
			return nil, err
		}
	}
}

func findProcessByName(name string) (*WindowsProcess, error) {
	ps, err := processes()
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		if strings.ToLower(p.Exe) == strings.ToLower(name) {
			return &p, nil
		}
	}
	return nil, errors.New("Cant's find specific process by name" + name)
}

func newWindowsProcess(e *windows.ProcessEntry32) WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return WindowsProcess{
		ProcessID:       int(e.ProcessID),
		ParentProcessID: int(e.ParentProcessID),
		Exe:             syscall.UTF16ToString(e.ExeFile[:end]),
	}
}

func GetPID(processName string) (int32, error){
	proc, err :=findProcessByName(processName)
	if err!=nil{
		logger.LogError("Get PID failed with error %s", err.Error())
		return -1, err
	}

	return int32(proc.ProcessID), nil
}

func OpenProcessHandle(processName string) (*syscall.Handle, error) {
	wp, err := findProcessByName(processName)
	if err != nil {
		logger.LogError("Find process ID failed with error %s", err.Error())
		return nil, err
	}

	h, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(wp.ProcessID))
	if err != nil{
		logger.LogError("Open process handle fail with error %s", err)
		return nil, err
	}

	return &h, nil
}

func GetAPI(dllName string, funcName string) (*syscall.LazyProc,error) {
	entry := syscall.NewLazyDLL(dllName)
	if entry == nil {
		logger.LogError("Can't find system library %s", dllName)
		return nil, errors.New("Can't find system library" + dllName)
	}

	proc :=entry.NewProc(funcName)
	if proc == nil{
		logger.LogError("Can't find function name %s in library %s", dllName)
		return nil, errors.New("Can't find function name"+ funcName + "library" + dllName)
	}
	return proc, nil
}
