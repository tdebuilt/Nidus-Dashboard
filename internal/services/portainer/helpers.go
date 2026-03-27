package portainer

import (
	"net"
	"net/url"
	"strings"
)

// ExtractHost extracts the hostname/IP from a URL.
func ExtractHost(rawURL string) string {
	if strings.HasPrefix(rawURL, "unix://") {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

// ResolveToIP resolves a hostname to an IP address. Returns the hostname as-is if already an IP.
func ResolveToIP(hostname string) string {
	if hostname == "" {
		return ""
	}
	if ip := net.ParseIP(hostname); ip != nil {
		return hostname
	}
	ips, err := net.LookupHost(hostname)
	if err != nil || len(ips) == 0 {
		return hostname
	}
	return ips[0]
}
