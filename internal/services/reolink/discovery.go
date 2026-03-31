package reolink

import (
	"encoding/xml"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

const onvifDiscoveryAddr = "239.255.255.250:3702"

const wsDiscoveryProbe = `<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope"
            xmlns:w="http://schemas.xmlsoap.org/ws/2004/08/addressing"
            xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"
            xmlns:dn="http://www.onvif.org/ver10/network/wsdl">
  <e:Header>
    <w:MessageID>uuid:nidus-discovery</w:MessageID>
    <w:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</w:To>
    <w:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</w:Action>
  </e:Header>
  <e:Body>
    <d:Probe>
      <d:Types>dn:NetworkVideoTransmitter</d:Types>
    </d:Probe>
  </e:Body>
</e:Envelope>`

var ipRegex = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)

type probeMatch struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		ProbeMatches struct {
			Match []struct {
				XAddrs string `xml:"XAddrs"`
				Scopes string `xml:"Scopes"`
			} `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}

// DiscoverCameras sends an ONVIF WS-Discovery probe and returns found cameras.
func DiscoverCameras(timeout time.Duration) ([]DiscoveredCamera, error) {
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	addr, err := net.ResolveUDPAddr("udp4", onvifDiscoveryAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve multicast addr: %w", err)
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	if _, err := conn.WriteTo([]byte(wsDiscoveryProbe), addr); err != nil {
		return nil, fmt.Errorf("send probe: %w", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	cameras := collectProbeResponses(conn)
	if cameras == nil {
		cameras = []DiscoveredCamera{}
	}
	return cameras, nil
}

// collectProbeResponses reads and parses WS-Discovery probe responses from the connection.
func collectProbeResponses(conn *net.UDPConn) []DiscoveredCamera {
	seen := make(map[string]bool)
	var cameras []DiscoveredCamera
	buf := make([]byte, 8192)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		var pm probeMatch
		if err := xml.Unmarshal(buf[:n], &pm); err != nil {
			continue
		}
		for _, match := range pm.Body.ProbeMatches.Match {
			ip := extractIP(match.XAddrs)
			if ip == "" || seen[ip] {
				continue
			}
			seen[ip] = true
			name, model := parseScopes(match.Scopes)
			cameras = append(cameras, DiscoveredCamera{IP: ip, Name: name, Model: model})
		}
	}
	return cameras
}

func extractIP(xaddrs string) string {
	match := ipRegex.FindString(xaddrs)
	return match
}

func parseScopes(scopes string) (name, model string) {
	parts := strings.Fields(scopes)
	for _, s := range parts {
		if strings.Contains(s, "onvif://www.onvif.org/name/") {
			name = strings.TrimPrefix(s, "onvif://www.onvif.org/name/")
			name = strings.ReplaceAll(name, "%20", " ")
		}
		if strings.Contains(s, "onvif://www.onvif.org/hardware/") {
			model = strings.TrimPrefix(s, "onvif://www.onvif.org/hardware/")
			model = strings.ReplaceAll(model, "%20", " ")
		}
	}
	if name == "" {
		name = "Unknown Camera"
	}
	return
}
