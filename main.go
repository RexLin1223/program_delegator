package main

import (
	"fmt"
	"scp_delegator/config"
	"scp_delegator/logger"
	"scp_delegator/task"
)

func main() {
	fmt.Println("Main start")
	logger.LogInfo("Main Start")
	//if mode != local_mode {
	// Check parent sign
	//}


	cfg := config.ParseProfile()
	mgr, err := task.CreateTaskHandler(cfg)
	if err != nil{
		logger.LogError("Can't initialize tasks manager with error %s", err.Error())
		return
	}

	mgr.Run()

	fmt.Println("Main End")
	logger.LogInfo("Main End")
}
