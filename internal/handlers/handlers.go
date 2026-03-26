package handlers

import "regexp"

// Package-level compiled regexes for error sanitization.
var (
	ipRegex   = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(:\d+)?`)
	dialRegex = regexp.MustCompile(`dial tcp ([^:]+):`)
)

// sanitizeError strips IP addresses and hostnames from error messages
// to avoid leaking internal network details in API responses.
func sanitizeError(err error) string {
	msg := err.Error()
	msg = ipRegex.ReplaceAllString(msg, "[redacted]")
	msg = dialRegex.ReplaceAllString(msg, "dial tcp [redacted]:")
	return msg
}
