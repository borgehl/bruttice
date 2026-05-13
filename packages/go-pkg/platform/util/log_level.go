package util

import "os"

var validLogLevels = map[string]bool{
	"debug":  true,
	"info":   true,
	"warn":   true,
	"error":  true,
	"silent": true,
}

func GetLogLevel(envVarName string) string {
	level := os.Getenv(envVarName)
	if !validLogLevels[level] {
		return "info"
	}
	return level
}
