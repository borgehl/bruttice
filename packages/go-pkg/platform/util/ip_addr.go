package util

import (
	"net/http"
	"strings"
)

func GetClientIPAddr(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		ips := strings.Split(r.Header.Get(header), ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}
	ipPort := strings.Split(r.RemoteAddr, ":")
	if len(ipPort) > 0 {
		return ipPort[0]
	}
	return ""
}
