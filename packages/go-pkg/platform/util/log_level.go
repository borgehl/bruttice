package util

import "os"

var validLogLevels = map[string]bool{
	"debug":  true,
	"info":   true,
	"warn":   true,
	"error":  true,
	"silent": true,
}

// GetLogLevel reads the named env var and returns it if it is a recognised
// level ("debug", "info", "warn", "error"). Defaults to "info".
func GetLogLevel(envVarName string) string {
	level := os.Getenv(envVarName)
	if !validLogLevels[level] {
		return "info"
	}
	return level
}
