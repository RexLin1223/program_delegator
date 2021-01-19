package main

import (
	"fmt"
	"scp_delegator/config"
	"scp_delegator/logger"
)

func main() {
	fmt.Println("Main start")
	logger.LogInfo("Main Start")
	if mode != local_mode {
		// Check parent sign
	}

	cfg := config.ParseProfile()
	if cfg != nil {
		for task := range cfg.Tasks {

		}
	}

	fmt.Println("Main End")
	logger.LogInfo("Main End")
}
