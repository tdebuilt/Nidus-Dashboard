package security

import (
	"strings"
	"testing"
)

func TestValidateExternalURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		// Valid URLs (use private IPs to avoid DNS dependency)
		{name: "valid http private IP", url: "http://192.168.1.100", wantErr: false},
		{name: "valid https private IP", url: "https://10.0.0.1", wantErr: false},
		{name: "valid http with port", url: "http://192.168.1.100:8080", wantErr: false},
		{name: "valid http with path", url: "http://172.16.0.1/api/v1", wantErr: false},

		// Blocked schemes
		{name: "ftp scheme", url: "ftp://example.com", wantErr: true, errMsg: "not allowed"},
		{name: "file scheme", url: "file:///etc/passwd", wantErr: true, errMsg: "not allowed"},
		{name: "javascript scheme", url: "javascript:alert(1)", wantErr: true, errMsg: "not allowed"},
		{name: "data scheme", url: "data:text/html,<h1>test</h1>", wantErr: true, errMsg: "not allowed"},

		// Empty / malformed
		{name: "empty URL", url: "", wantErr: true},
		{name: "no scheme", url: "example.com", wantErr: true},
		{name: "no hostname http", url: "http://", wantErr: true, errMsg: "no hostname"},

		// Blocked CIDRs (loopback)
		{name: "loopback 127.0.0.1", url: "http://127.0.0.1", wantErr: true, errMsg: "blocked IP"},
		{name: "loopback 127.0.0.2", url: "http://127.0.0.2", wantErr: true, errMsg: "blocked IP"},
		{name: "localhost", url: "http://localhost", wantErr: true, errMsg: "blocked IP"},

		// Blocked CIDRs (link-local / cloud metadata)
		{name: "link-local 169.254.169.254", url: "http://169.254.169.254", wantErr: true, errMsg: "blocked IP"},
		{name: "link-local 169.254.0.1", url: "http://169.254.0.1", wantErr: true, errMsg: "blocked IP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExternalURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for URL %q, got nil", tt.url)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for URL %q: %v", tt.url, err)
				}
			}
		})
	}
}

func TestParsedBlockedNetsInitialized(t *testing.T) {
	t.Parallel()
	if len(parsedBlockedNets) != len(blockedCIDRs) {
		t.Errorf("expected %d blocked nets, got %d", len(blockedCIDRs), len(parsedBlockedNets))
	}
}
