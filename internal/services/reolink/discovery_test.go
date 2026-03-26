package reolink

import (
	"testing"
)

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name   string
		xaddrs string
		want   string
	}{
		{
			name:   "valid XAddrs with IP",
			xaddrs: "http://192.168.1.100:2020/onvif/device_service",
			want:   "192.168.1.100",
		},
		{
			name:   "no IP in string",
			xaddrs: "http://localhost/onvif/device_service",
			want:   "",
		},
		{
			name:   "multiple IPs returns first",
			xaddrs: "http://10.0.0.1/path http://10.0.0.2/path",
			want:   "10.0.0.1",
		},
		{
			name:   "empty string",
			xaddrs: "",
			want:   "",
		},
		{
			name:   "IP only",
			xaddrs: "192.168.0.55",
			want:   "192.168.0.55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIP(tt.xaddrs)
			if got != tt.want {
				t.Errorf("extractIP(%q) = %q, want %q", tt.xaddrs, got, tt.want)
			}
		})
	}
}

func TestParseScopes(t *testing.T) {
	tests := []struct {
		name      string
		scopes    string
		wantName  string
		wantModel string
	}{
		{
			name:      "name and model present",
			scopes:    "onvif://www.onvif.org/type/NetworkVideoTransmitter onvif://www.onvif.org/name/RLC-810A onvif://www.onvif.org/hardware/RLC-810A",
			wantName:  "RLC-810A",
			wantModel: "RLC-810A",
		},
		{
			name:      "only name",
			scopes:    "onvif://www.onvif.org/name/FrontDoor onvif://www.onvif.org/type/video",
			wantName:  "FrontDoor",
			wantModel: "",
		},
		{
			name:      "only model",
			scopes:    "onvif://www.onvif.org/hardware/E1 onvif://www.onvif.org/type/video",
			wantName:  "Unknown Camera",
			wantModel: "E1",
		},
		{
			name:      "neither name nor model",
			scopes:    "onvif://www.onvif.org/type/NetworkVideoTransmitter",
			wantName:  "Unknown Camera",
			wantModel: "",
		},
		{
			name:      "URL-encoded spaces in name",
			scopes:    "onvif://www.onvif.org/name/Front%20Door%20Camera onvif://www.onvif.org/hardware/RLC%20811A",
			wantName:  "Front Door Camera",
			wantModel: "RLC 811A",
		},
		{
			name:      "empty scopes",
			scopes:    "",
			wantName:  "Unknown Camera",
			wantModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotModel := parseScopes(tt.scopes)
			if gotName != tt.wantName {
				t.Errorf("parseScopes(%q) name = %q, want %q", tt.scopes, gotName, tt.wantName)
			}
			if gotModel != tt.wantModel {
				t.Errorf("parseScopes(%q) model = %q, want %q", tt.scopes, gotModel, tt.wantModel)
			}
		})
	}
}
