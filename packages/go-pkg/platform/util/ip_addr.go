package util

import (
	"net/http"
	"strings"
)

// GetClientIPAddr extracts the client's IP address from the request.
func GetClientIPAddr(r *http.Request) string {
	// Check common headers for forwarded IP addresses.
	// The X-Forwarded-For header can contain multiple IPs (client, proxy1, proxy2),
	// so we take the first one, which is typically the original client's IP.
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		ips := strings.Split(r.Header.Get(header), ",")
		if len(ips) > 0 {
			// Trim spaces and return the first IP address found
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}
	// Fallback to RemoteAddr if no proxy headers are found.
	// RemoteAddr includes both IP and port (e.g., "127.0.0.1:8080").
	// We split to get only the IP part.
	ipPort := strings.Split(r.RemoteAddr, ":")
	if len(ipPort) > 0 {
		return ipPort[0]
	}
	return ""
}
