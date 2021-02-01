package constant

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

// EmbedBinaryOptions is a const map used to validate specified executable name given in profile.
var EmbedBinaryOptions = map[string]string{
	"wpr.exe":             `-w`,
	"toolDiskVolScan.exe": `-d`,
	"dsa_control.cmd":     `-c`,
	"dsa_query.cmd":       `-q`,
	"sendCommand.cmd":     `-s`,
	"ratt.exe":            `-r`,
}
