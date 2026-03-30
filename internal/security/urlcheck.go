package security

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
)

// Subnets that must be blocked to prevent SSRF attacks.
var blockedCIDRs = []string{
	"127.0.0.0/8",    // loopback
	"169.254.0.0/16", // link-local / cloud metadata
	"::1/128",        // IPv6 loopback
}

// parsedBlockedNets is initialised once from blockedCIDRs.
var parsedBlockedNets []*net.IPNet

func init() {
	for _, cidr := range blockedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Fatalf("security: invalid blocked CIDR %q: %v", cidr, err)
		}
		parsedBlockedNets = append(parsedBlockedNets, ipNet)
	}
}

// ValidateExternalURL checks that rawURL is safe to fetch.
// It rejects non-HTTP(S) schemes and dangerous destination IPs
// (loopback, link-local/cloud metadata) while allowing private
// network ranges (10.x, 192.168.x, 172.16-31.x) that are
// expected in a self-hosted environment.
func ValidateExternalURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme %q is not allowed, only http and https are accepted", u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("URL has no hostname")
	}

	ips, err := net.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("DNS lookup failed for %q: %w", hostname, err)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return fmt.Errorf("could not parse resolved IP %q", ipStr)
		}

		for _, blocked := range parsedBlockedNets {
			if blocked.Contains(ip) {
				return fmt.Errorf("hostname %q resolves to blocked IP %s (SSRF protection)", hostname, ipStr)
			}
		}
	}

	return nil
}
