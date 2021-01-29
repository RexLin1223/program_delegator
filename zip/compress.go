package zip

import (
	"context"
	"errors"
	"fmt"
	"github.com/beevik/guid"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"scp_delegator/config"
	"scp_delegator/constant"
	"scp_delegator/logger"
)

// CompressLogDir will use 7z tool to compress directory/file into .7z/.zip file.
func CompressLogDir(ctx *context.Context, cfg *config.Upload, logDir string) ([]string, error) {
	toolPath := getZipToolPath()
	if toolPath == "" {
		return nil, errors.New("can't get zip tool path")
	}

	filePath := getOutputFilePath()
	if filePath == "" {
		return nil, errors.New("can't get compress file output path")
	}



	// Compose exec command
	cmd := exec.CommandContext(*ctx, toolPath, "a", filePath, logDir,
		fmt.Sprintf("-v%dm", cfg.MaxBlockSizeMB), "-mmt1")
	logger.Wrapper.LogTrace("Exec compressing with command: %s", cmd.String())

	// Close logger to prevent file lock in compressing.
	logger.Wrapper.Close()
	s, err := cmd.Output()
	logger.Wrapper.LogTrace("Compressing result: %s", s)
	if err != nil {
		return nil, err
	}

	return getCompressedSegments(filePath)
}

func getCompressedSegments(compressedFilePath string) ([]string, error) {
	pattern := compressedFilePath + ".*"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func getOutputFilePath() string {
	p, err := os.Getwd()
	if err != nil {
		return ""
	}
	fileName := guid.NewString() +"_" +constant.CompressedFileName
	return filepath.Join(p, fileName)
}

func getZipToolPath() string {
	p, err := os.Getwd()
	if err != nil {
		return ""
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(p, constant.WindowsZipToolBinary)
	} else if runtime.GOOS == "linux" {
		// TODO: Assign linux tgz tool
		return ""
	} else {
		return ""
	}
}
