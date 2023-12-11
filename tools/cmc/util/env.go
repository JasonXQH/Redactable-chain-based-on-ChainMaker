package util

import "os"

const (
	envKeySdkConfPath = "CMC_SDK_CONF_PATH"
)

var (
	EnvSdkConfPath string
)

func init() {
	EnvSdkConfPath = os.Getenv(envKeySdkConfPath)
}
