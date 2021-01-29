package constant

// EmbedBinaryOptions is a const map used to validate specified executable name given in profile.
const (
	ExecutorName       = "rp_main.exe"
	OutputDirectory    = "Log"
	LogFileName        = "ds_scp.log"
	CompressedFileName = "log.7z"
)

// Windows
const (
	WindowsXBCLogDir32   = "C:\\Program Files (x86)\\Trend Micro\\Endpoint Basecamp\\log"
	WindowsXBCLogDir64   = "C:\\Program Files (x86)\\Trend Micro\\Endpoint Basecamp\\log"
	WindowsZipToolBinary = "7z.exe"
)

// Linux
const (
	LinuxXBCLogDir32 = ""
	LinuxXBCLogDir64 = ""
)

var EmbedBinaryOptions = map[string]string{
	"wpr.exe":             `-w`,
	"toolDiskVolScan.exe": `-d`,
	"dsa_control.cmd":     `-t "dsa_control"`,
	"dsa_query.cmd":       `-t "dsa_query"`,
	"sendCommand.cmd":     `-t "sendCommand"`,
	"ratt.exe":            `-t "ratt"`,
}
