package metric

import "scp_delegator/logger"

func main() {
	i32, err := GetCoreCounts(false)
	logger.Wrapper.LogInfo("Core count %d, error=%s", i32, err)

	f64, err := GetCpuUsage()
	logger.Wrapper.LogInfo("CPU usage %f percent, error=%s", f64, err)
	i64, err := GetMemoryUsageMB()
	logger.Wrapper.LogInfo("Total memory usage %d MB, error=%s", i64, err)

	// Get specific process
	proceeName := "TaskMgr.exe"
	f64, err = GetProcessCpuUsage(proceeName)
	logger.Wrapper.LogInfo("Get process CPU usage %f percent, error=%s", f64, err)

	i64, err = GetProcessMemoryUsageByte(proceeName)
	logger.Wrapper.LogInfo("Get Process %s, memory usage %d Byte, error=%s", proceeName, i64, err)

	i64, err = GetProcessMemoryUsageMB(proceeName)
	logger.Wrapper.LogInfo("Get Process %s, memory usage %d MB, error=%s", proceeName, i64, err)
}
