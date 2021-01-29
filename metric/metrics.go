package metric

import (
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	win "scp_delegator/system/windows"
	"syscall"
	"time"
	"unsafe"
)

func GetCoreCounts(logical bool) (int32, error) {
	counts, err := cpu.Counts(logical)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Error occurs when get CPU core counts, error=%d", counts))
	}
	return int32(counts), nil
}

// GetCpuUsage will get CPU snapshot
func GetCpuUsage() (float64,error) {
	percent, err := cpu.Percent(10*time.Millisecond, false)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when get CPU usage, %s", err.Error()))
	}
	return percent[0], nil
}

func GetCpuUsageCores() ([]float64, error) {
	var ps []float64
	ps, err := cpu.Percent(1000*time.Millisecond, false)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error occurs when get CPU usage, %s", err.Error()))
	}
	return ps, nil
}

func GetProcessCpuUsage(processName string) (float64,error) {
	pid, err := win.GetPID(processName)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when query PID with process %s, error=%s", processName, err.Error()))
	}
	proc, err := process.NewProcess(pid)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when create process, error=%s", err.Error()))
	}
	percent, err := proc.CPUPercent()
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when get process CPU usage, error=%s", err.Error()))
	}
	return percent, nil
}

func GetMemoryUsageByte() (int64,error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Error occurs when get memory usage, error=%s", err.Error()))
	}
	return int64(v.Used),nil
}

type PROCESS_MEMORY_COUNTERS_EX struct {
	cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uint64
	WorkingSetSize             uint64
	QuotaPeakPagedPoolUsage    uint64
	QuotaPagedPoolUsage        uint64
	QuotaPeakNonPagedPoolUsage uint64
	QuotaNonPagedPoolUsage     uint64
	PagefileUsage              uint64
	PeakPagefileUsage          uint64
	PrivateUsage               uint64
}

func GetProcessMemoryUsageByte(processName string) (int64, error) {
	procHandle, err := win.OpenProcessHandle(processName)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when open process %s failed, error=%s", processName, err.Error()))
	}
	defer syscall.CloseHandle(*procHandle)

	proc, err := win.GetAPI("Psapi.dll", "GetProcessMemoryInfo")
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when get API failed, error=%s", err.Error()))
	}

	pmc := PROCESS_MEMORY_COUNTERS_EX{}
	_, _, err = proc.Call(uintptr(*procHandle), uintptr(unsafe.Pointer(&pmc)), unsafe.Sizeof(pmc))
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error occurs when calling Windows API GetProcessMemoryInfo() failed, error %s", err.Error()))
	}

	return (int64)(pmc.PrivateUsage), nil
}

func GetProcessMemoryUsageMB(processName string) (int64,error) {
	m, err := GetProcessMemoryUsageByte(processName)
	if err != nil {
		return -1, err
	}
	return m >> 20, nil
}

func GetMemoryUsageMB() (int64,error) {
	m,err := GetMemoryUsageByte()
	if err!=nil{
		return -1,err
	}
	return m >> 20, nil
}

func getDiskUsageStat(path string) (*disk.UsageStat, error) {
	usageStat, err := disk.Usage(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Get disk usage failed with error %s", err.Error()))
	}
	return usageStat, nil
}

func GetDiskUsageGB(path string) (int64,error) {
	usageStat,err := getDiskUsageStat(path)
	if err != nil {
		return -1, err
	}

	return int64(usageStat.Used >> 30), nil
}

func GetDiskFreeGB(path string) (int64,error) {
	usageStat,err := getDiskUsageStat(path)
	if err != nil {
		return -1, err
	}

	return int64(usageStat.Free >> 30), nil
}

func GetDiskTotalGB(path string) (int64, error) {
	usageStat,err := getDiskUsageStat(path)
	if err!=nil {
		return -1, err
	}

	return int64(usageStat.Total >> 30), nil
}
