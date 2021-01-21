package config

// EmbbedBinaryList is a const map used to validate specified executalbe name given in profile.
var EmbbedBinaryList = map[string]bool{
	"wpr.exe":             true,
	"toolDiskVolScan.exe": true,
	"dsa_control.cmd":     true,
	"dsa_query.cmd":       true,
	"sendCommand.cmd":     true,
	"ratt.exe":            true,
}
