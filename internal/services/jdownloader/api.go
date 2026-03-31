package jdownloader

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// callAction calls a device action through the MyJDownloader cloud relay.
// It auto-connects if no session exists and retries once on auth failure.
func (c *Client) callAction(ctx context.Context, action string, params []any, result any) error {
	if err := c.ensureConnected(ctx); err != nil {
		return err
	}

	err := c.doDeviceCall(ctx, action, params, result)
	if err == nil {
		return nil
	}

	// Retry once after reconnect
	c.mu.Lock()
	c.sessionToken = ""
	c.deviceID = ""
	c.mu.Unlock()

	if err := c.ensureConnected(ctx); err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}
	return c.doDeviceCall(ctx, action, params, result)
}

func (c *Client) doDeviceCall(ctx context.Context, action string, params []any, result any) error {
	c.mu.Lock()
	session := c.sessionToken
	device := c.deviceID
	devToken := c.deviceEncryptionToken
	c.mu.Unlock()

	req, err := c.buildEncryptedRequest(ctx, session, device, action, params, devToken)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if ciphertext, err := base64.StdEncoding.DecodeString(string(respBytes)); err == nil {
			if plaintext, err := decrypt(ciphertext, devToken); err == nil {
				return fmt.Errorf("status %d: %s", resp.StatusCode, string(plaintext))
			}
		}
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(respBytes))
	}

	if result == nil {
		return nil
	}
	return c.decryptDeviceResponse(respBytes, devToken, result)
}

// buildEncryptedRequest constructs an encrypted HTTP request for a JDownloader device action.
func (c *Client) buildEncryptedRequest(ctx context.Context, session, device, action string, params []any, devToken []byte) (*http.Request, error) {
	rid := c.nextRid()
	actionPath := "/" + action
	httpPath := fmt.Sprintf("/t_%s_%s%s",
		url.QueryEscape(session), url.QueryEscape(device), actionPath)

	body := map[string]any{
		"url":    actionPath,
		"rid":    rid,
		"apiVer": 1,
	}
	if params != nil {
		body["params"] = params
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling body: %w", err)
	}

	encBody, err := encrypt(jsonBody, devToken)
	if err != nil {
		return nil, fmt.Errorf("encrypting body: %w", err)
	}

	reqURL := apiURL + httpPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(base64.StdEncoding.EncodeToString(encBody)))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/aesjson-jd; charset=utf-8")
	return req, nil
}

func (c *Client) decryptDeviceResponse(respBytes, token []byte, result any) error {
	ciphertext, err := base64.StdEncoding.DecodeString(string(respBytes))
	if err != nil {
		return fmt.Errorf("base64 decoding response: %w", err)
	}

	plaintext, err := decrypt(ciphertext, token)
	if err != nil {
		return fmt.Errorf("decrypting response: %w", err)
	}

	var apiResp struct {
		Data json.RawMessage `json:"data"`
		Rid  int64           `json:"rid"`
	}
	if err := json.Unmarshal(plaintext, &apiResp); err != nil {
		return fmt.Errorf("decoding response JSON: %w", err)
	}

	if apiResp.Data != nil {
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("decoding data: %w", err)
		}
	}
	return nil
}

// callServer makes a signed server call (connect, listdevices, disconnect).
func (c *Client) callServer(ctx context.Context, query string, secret []byte, result any) error {
	rid := c.nextRid()
	query += fmt.Sprintf("&rid=%d", rid)
	signature := sign(secret, query)
	query += "&signature=" + signature

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+query, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	if result == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return fmt.Errorf("base64 decoding: %w", err)
	}

	plaintext, err := decrypt(ciphertext, secret)
	if err != nil {
		return fmt.Errorf("decrypting response: %w", err)
	}

	if err := json.Unmarshal(plaintext, result); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}
	return nil
}
