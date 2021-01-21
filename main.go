package main

import (
	"fmt"
	"scp_delegator/config"
	"scp_delegator/logger"
	"scp_delegator/metric"
)

func main() {
	fmt.Println("Main start")
	logger.LogInfo("Main Start")
	//if mode != local_mode {
	// Check parent sign
	//}

	logger.LogInfo("Core count %d", metric.GetCoreCounts(false))
	logger.LogInfo("CPU usage %f", metric.GetCpuUsage())
	logger.LogInfo("Total memory usage %d MB", metric.GetMemoryUsageMB())
	logger.LogInfo("Get Process memory count %d Byte", metric.GetProcessMemoryUsageByte("Zoom.exe"))

	cfg := config.ParseProfile()
	if cfg != nil {
		for _, task := range cfg.Tasks {
			fmt.Printf("Run task ID= %d, name= %s.\n", task.ID, task.Name)
			logger.LogInfo("Run task ID= %d, name= %s.\n", task.ID, task.Name)

		}
	}

	fmt.Println("Main End")
	logger.LogInfo("Main End")
}
