package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Sender sends notifications through various providers.
type Sender struct {
	httpClient *http.Client
}

// NewSender creates a new notification sender.
func NewSender() *Sender {
	return &Sender{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches a notification to the given provider.
func (s *Sender) Send(ctx context.Context, providerType, url, token, config, title, message string) error {
	switch providerType {
	case "gotify":
		return s.sendGotify(ctx, url, token, title, message)
	case "ntfy":
		return s.sendNtfy(ctx, url, config, title, message)
	case "apprise":
		return s.sendApprise(ctx, url, config, title, message)
	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// sendGotify sends a notification via Gotify.
// POST {url}/message?token={token}
func (s *Sender) sendGotify(ctx context.Context, baseURL, token, title, message string) error {
	baseURL = strings.TrimRight(baseURL, "/")
	endpoint := fmt.Sprintf("%s/message?token=%s", baseURL, token)

	body, err := json.Marshal(map[string]any{
		"title":    title,
		"message":  message,
		"priority": 5,
	})
	if err != nil {
		return fmt.Errorf("marshal gotify payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gotify request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gotify request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gotify returned status %d", resp.StatusCode)
	}
	return nil
}

// sendNtfy sends a notification via Ntfy.
// POST {url}/{topic}
func (s *Sender) sendNtfy(ctx context.Context, baseURL, configJSON, title, message string) error {
	baseURL = strings.TrimRight(baseURL, "/")

	var cfg struct {
		Topic string `json:"topic"`
	}
	if configJSON != "" && configJSON != "{}" {
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return fmt.Errorf("invalid ntfy config: %w", err)
		}
	}
	if cfg.Topic == "" {
		cfg.Topic = "nidus"
	}

	endpoint := fmt.Sprintf("%s/%s", baseURL, cfg.Topic)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("ntfy request creation failed: %w", err)
	}
	req.Header.Set("Title", title)
	req.Header.Set("Priority", "default")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}
	return nil
}

// sendApprise sends a notification via Apprise API.
// POST {url}/notify
func (s *Sender) sendApprise(ctx context.Context, baseURL, configJSON, title, message string) error {
	baseURL = strings.TrimRight(baseURL, "/")

	var cfg struct {
		URLs []string `json:"urls"`
	}
	if configJSON != "" && configJSON != "{}" {
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return fmt.Errorf("invalid apprise config: %w", err)
		}
	}

	payload := map[string]any{
		"title": title,
		"body":  message,
		"type":  "info",
	}
	if len(cfg.URLs) > 0 {
		payload["urls"] = strings.Join(cfg.URLs, ",")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal apprise payload: %w", err)
	}
	endpoint := fmt.Sprintf("%s/notify", baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("apprise request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("apprise request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("apprise returned status %d", resp.StatusCode)
	}
	return nil
}
