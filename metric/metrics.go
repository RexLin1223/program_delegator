package metric

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"scp_delegator/logger"
	win "scp_delegator/os/windows"
	"syscall"
	"time"
	"unsafe"
)

func GetCoreCounts(logical bool) uint32 {
	counts, err := cpu.Counts(logical)
	if err != nil {
		logger.LogError("Error occurs when get CPU core counts, %s", counts)
		return 0
	}
	return uint32(counts)
}

// GetCpuUsage will get CPU snapshot
func GetCpuUsage() float64 {
	percent, err := cpu.Percent(10*time.Millisecond, false)
	if err != nil {
		logger.LogError("Error occurs when get CPU usage, %s", err.Error())
		return -1
	}
	return percent[0]
}

func GetCpuUsageCores() []float64 {
	var ps []float64
	ps, err := cpu.Percent(1000*time.Millisecond, false)
	if err != nil {
		logger.LogError("Error occurs when get CPU usage, %s", err.Error())
	}
	return ps
}

func GetProcessCpuUsage(processName string) float64 {
	pid, err := win.GetPID(processName)
	if err != nil {
		logger.LogError("Error occurs when query PID with process %s, error=%s", processName, err.Error())
		return -1
	}
	proc, err := process.NewProcess(pid)
	if err != nil {
		logger.LogError("Error occurs when create process, error=%s", err.Error())
		return -1
	}
	percent, err := proc.CPUPercent()
	if err != nil {
		logger.LogError("Error occurs when get process CPU usage")
		return -1
	}
	return percent
}

func GetMemoryUsageByte() uint64 {
	v, err := mem.VirtualMemory()
	if err != nil {
		logger.LogError("Error occurs when get memory usage", err.Error())
	}
	return v.Used
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

func GetProcessMemoryUsageByte(processName string) int64 {
	procHandle, err := win.OpenProcessHandle(processName)
	if err != nil {
		logger.LogError("Get process memory fail with open process fail")
		return -1
	}
	defer syscall.CloseHandle(*procHandle)

	proc, err := win.GetAPI("Psapi.dll", "GetProcessMemoryInfo")
	if err != nil {
		logger.LogError("Get process memory fail with get API fail")
		return -1
	}

	pmc := PROCESS_MEMORY_COUNTERS_EX{}
	_, _, err = proc.Call(uintptr(unsafe.Pointer(*procHandle)), uintptr(unsafe.Pointer(&pmc)), unsafe.Sizeof(pmc))
	if err != nil {
		logger.LogError("Get memory by API GetProcessMemoryInfo fail with error %s", err.Error())
	}

	return (int64)(pmc.PrivateUsage)
}

func GetProcessMemoryUsageMB(processName string) int64 {
	m := GetProcessMemoryUsageByte(processName)
	if m == -1 {
		return -1
	}
	return m >> 20
}

func GetMemoryUsageMB() uint64 {
	m := GetMemoryUsageByte()
	return m >> 20
}

func getDiskUsageStat(path string) *disk.UsageStat{
	usageStat, err:= disk.Usage(path)
	if err != nil{
		logger.LogError("Get disk usage failed with error %s", err.Error())
		return nil
	}
	return usageStat
}

func GetDiskUsageGB(path string) int64 {
	usageStat := getDiskUsageStat(path)
	if usageStat == nil{
		return -1
	}

	return int64(usageStat.Used >> 30)
}

func GetDiskFreeGB(path string) int64 {
	usageStat := getDiskUsageStat(path)
	if usageStat == nil{
		return -1
	}

	return int64(usageStat.Free >> 30)
}


func GetDiskTotalGB(path string) int64 {
	usageStat := getDiskUsageStat(path)
	if usageStat == nil{
		return -1
	}

	return int64(usageStat.Total >> 30)
}