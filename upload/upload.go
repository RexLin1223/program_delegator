package upload

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"scp_delegator/config"
	"scp_delegator/logger"
	"time"
)

type Result struct {
	Error    *error
	Response *http.Response
}

// PutFileToAzBlob will upload file to Azure blob
func PutFileToAzBlob(ctx *context.Context, cfg *config.Upload, filePath string) *Result {
	uploadCtx, cancelFunc := context.WithTimeout(*ctx, time.Second*time.Duration(cfg.TimeoutS))
	defer cancelFunc()

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return &Result{
			Error:    &err,
			Response: nil,
		}
	}
	defer file.Close()

	// Set proxy to env
	setProxy(cfg)

	// Prepare upload URL, pipeline, and options
	/*
		retryOpts := azblob.RetryOptions{
			Policy:                      azblob.RetryPolicyFixed,
			MaxTries:                    int32(cfg.MaxRetryCount),
			//TryTimeout:                  time.Duration(cfg.TimeoutS),
			RetryDelay:                  0,
			MaxRetryDelay:               0,
			RetryReadsFromSecondaryHost: "",
		}
	*/

	pipe := azblob.NewPipeline(azblob.NewAnonymousCredential(), azblob.PipelineOptions{})
	url, err := url.Parse(composeAzBlobURL(cfg, filePath))
	if err != nil {
		return &Result{
			Error:    &err,
			Response: nil,
		}
	}

	blobURL := azblob.NewBlockBlobURL(*url, pipe)
	blobOptions := azblob.UploadToBlockBlobOptions{
		BlockSize:   int64(cfg.MaxBlockSizeMB << 10),
		Parallelism: 16,
	}

	response, err := azblob.UploadFileToBlockBlob(uploadCtx, file, blobURL, blobOptions)
	if err != nil {
		return &Result{
			Error:    &err,
			Response: nil,
		}
	}

	return &Result{
		Response: response.Response(),
		Error:    nil,
	}
}

// PutFiles uploads each files in blocks to a block blob.
func PutFilesToAzBlob(ctx *context.Context, cfg *config.Upload, filesPath []string) (results []*Result) {
	for _, f := range filesPath {
		r := PutFileToAzBlob(ctx, cfg, f)
		results = append(results, r)
	}

	return results
}

func setProxy(cfg *config.Upload) {
	if cfg.Proxy.Host != "" && cfg.Proxy.Port != 0 {
		s := fmt.Sprintf("http://%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
		if err := os.Setenv("HTTP_PROXY", s); err != nil {
			logger.Wrapper.LogError("Error when set environment http_proxy")
		}
		s = fmt.Sprintf("https://%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
		if err := os.Setenv("HTTPS_PROXY", s); err != nil {
			logger.Wrapper.LogError("Error when set environment https_proxy")
		}
	}
}

func composeAzBlobURL(cfg *config.Upload, filePath string) string {
	fileName := filepath.Base(filePath)
	return fmt.Sprintf("https://%s.%s/%s/%s/%s/%s/%s%s",
		cfg.AzBlob.AccountName, cfg.AzBlob.HostName,
		cfg.AzBlob.ContainerName, cfg.SEGCaseID,
		cfg.CompanyID, cfg.DeviceID, fileName,
		cfg.AzBlob.SASToken)
}

func progressUpdate(byteTransferred int64) {
	fmt.Printf("Transferred bytes %d", byteTransferred)
}
