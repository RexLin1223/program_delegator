package task

// EmbedBinaryOptions is a const map used to validate specified executable name given in profile.
var EmbedBinaryOptions = map[string]string{
	"wpr.exe":             `-w`,
	"toolDiskVolScan.exe": `-d`,
	"dsa_control.cmd":     `-t "dsa_control"`,
	"dsa_query.cmd":       `-t "dsa_query"`,
	"sendCommand.cmd":     `-t "sendCommand"`,
	"ratt.exe":            `-t "ratt"`,
}
