package metric

import "scp_delegator/logger"

func main() {
	logger.LogInfo("Core count %d", GetCoreCounts(false))
	logger.LogInfo("CPU usage %f percent", GetCpuUsage())
	logger.LogInfo("Get process CPU usage %f percent", GetProcessCpuUsage("TaskMgr.exe"))
	logger.LogInfo("Total memory usage %d MB", GetMemoryUsageMB())
	logger.LogInfo("Get Process memory usage %d Byte", GetProcessMemoryUsageByte("TaskMgr.exe"))
	logger.LogInfo("Get Process memory usage %d MB", GetProcessMemoryUsageMB("TaskMgr.exe"))
}