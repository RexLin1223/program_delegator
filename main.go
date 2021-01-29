package main

import (
	"context"
	"fmt"
	"scp_delegator/config"
	"scp_delegator/logger"
	"scp_delegator/task"
	"scp_delegator/upload"
	"scp_delegator/zip"
)

func main() {
	fmt.Println("Main start")
	logger.Wrapper.LogInfo("Main Start")

	//if mode != local_mode {
	// Check parent sign
	//}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cfg := config.ParseProfile()
	mgr, err := task.CreateTaskHandler(&ctx, cfg)
	if err != nil {
		logger.Wrapper.LogError("Can't initialize tasks manager with error %s", err.Error())
		return
	}

	// Executor
	mgr.Run()

	// Compress & upload Log dir
	compressAndUpload(&ctx, cfg, config.GetOutputDir())

	fmt.Println("Main End")
	logger.Wrapper.LogInfo("Main End")
	// Compressing and uploading final log
	compressAndUpload(&ctx, cfg, logger.GetOutputPath())
}

func compressAndUpload(ctx *context.Context, cfg *config.Config, target string) {
	// Compressing
	segments, err := zip.CompressLogDir(ctx, &cfg.Upload, target)
	if err != nil {
		logger.Wrapper.LogFatal("[Main] Get error when compressing output directory, error=%s", err)
		return
	}
	// Uploading
	results := upload.PutFilesToAzBlob(ctx, &cfg.Upload, segments)
	for _, r := range results {
		if r.Error != nil {
			logger.Wrapper.LogError("Upload failed with error=%s", r.Error)
		}
	}
}
