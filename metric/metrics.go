package metric

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

func GetCpuUsageCores() ([]float64, error) {
	var ps []float64
	ps, err := cpu.Percent(1000*time.Millisecond, false)
	if err != nil {
		logger.LogError("Error occurs when get CPU usage, %s", err.Error())
		return ps, err
	}
	return ps, nil
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
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
}

func GetProcessMemoryUsageByte(processName string) uint64 {
	procHandle := win.OpenProcessHandle(processName)
	if procHandle == nil {
		logger.LogError("Get process memory fail with open process fail")
		return 0
	}
	defer syscall.CloseHandle(*procHandle)

	proc := win.GetAPI("Psapi.dll", "GetProcessMemoryInfo")
	if proc == nil {
		logger.LogError("Get process memory fail with get API fail")
		return 0
	}

	pmc := PROCESS_MEMORY_COUNTERS_EX{}
	_, _, err := proc.Call(uintptr(unsafe.Pointer(*procHandle)), uintptr(unsafe.Pointer(&pmc)), unsafe.Sizeof(pmc))
	if err!=nil{
		logger.LogError("Get memory by API GetProcessMemoryInfo fail with error %s", err.Error())
	}
	// TODO unload dll

	return *(*uint64)(unsafe.Pointer(pmc.PrivateUsage))
}

func GetMemoryUsageMB() uint64 {
	m := GetMemoryUsageByte()
	return m >> 20
}
