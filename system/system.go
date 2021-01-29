package system

import (
	"runtime"
)

func GetOSandArch() string {
	return runtime.GOOS + runtime.GOARCH
}
