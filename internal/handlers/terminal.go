package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// TerminalHandler handles SSH terminal WebSocket proxy requests.
type TerminalHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

// terminalConfig is the SSH connection info stored in the widget config JSON.
type terminalConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	FontSize   int    `json:"font_size"`
	Scrollback int    `json:"scrollback"`
}

// terminalMsg is a message sent from the browser to the terminal WebSocket.
type terminalMsg struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint32 `json:"cols,omitempty"`
	Rows uint32 `json:"rows,omitempty"`
}

// HandleWS upgrades to WebSocket and proxies an SSH session.
//
// The WebSocket is upgraded before any SSH attempt so that connection,
// authentication and decryption errors can be reported back to the browser
// as structured error frames instead of opaque "WebSocket failed" failures.
func (h *TerminalHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	widgetID, ok := parseWidgetIDParam(w, r)
	if !ok {
		return
	}

	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("terminal: WebSocket upgrade failed", "error", err)
		return
	}
	defer clientConn.Close()

	cfg, err := h.loadTerminalConfig(r.Context(), widgetID)
	if err != nil {
		slog.Error("terminal: config error", "error", err, "widget_id", widgetID)
		sendWSError(clientConn, err.Error())
		return
	}

	sshClient, session, err := dialSSH(cfg)
	if err != nil {
		msg := classifyDialError(err, cfg.Host, cfg.Port)
		slog.Error("terminal: SSH dial failed",
			"error", err, "host", cfg.Host, "port", cfg.Port, "user", cfg.Username)
		sendWSError(clientConn, msg)
		return
	}
	defer sshClient.Close()
	defer session.Close()

	if err := h.runSSHSession(clientConn, session); err != nil {
		slog.Error("terminal: SSH session failed", "error", err)
		sendWSError(clientConn, err.Error())
	}
}

// runSSHSession wires the SSH session to the WebSocket and runs until either
// side closes. Returns an error if the session could not be started.
func (h *TerminalHandler) runSSHSession(clientConn *websocket.Conn, session *ssh.Session) error {
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("SSH session error: %w", err)
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("SSH session error: %w", err)
	}

	if err := session.RequestPty("xterm-256color", 24, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return fmt.Errorf("PTY request failed: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("shell start failed: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pumpSSHToWS(stdoutPipe, clientConn)
	}()

	readClientMessages(clientConn, stdinPipe, session)
	session.Close()
	wg.Wait()
	return nil
}

// pumpSSHToWS forwards SSH stdout bytes to the WebSocket as binary frames.
func pumpSSHToWS(stdoutPipe io.Reader, clientConn *websocket.Conn) {
	buf := make([]byte, 32*1024)
	for {
		n, readErr := stdoutPipe.Read(buf)
		if n > 0 {
			if writeErr := clientConn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				return
			}
		}
		if readErr != nil {
			return
		}
	}
}

// parseWidgetIDParam reads and validates the widget_id query parameter.
// On failure it writes a 400 response and returns ok=false.
func parseWidgetIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	widgetIDStr := r.URL.Query().Get("widget_id")
	if widgetIDStr == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "widget_id required"})
		return 0, false
	}
	widgetID, err := strconv.ParseInt(widgetIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid widget_id"})
		return 0, false
	}
	return widgetID, true
}

// sendWSError writes a JSON-encoded error frame to the WebSocket and closes it.
// The browser-side TerminalWidget renders these frames in red inside xterm.
func sendWSError(clientConn *websocket.Conn, msg string) {
	payload, err := json.Marshal(struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{Type: "error", Message: msg})
	if err != nil {
		return
	}
	_ = clientConn.WriteMessage(websocket.TextMessage, payload)
	_ = clientConn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseInternalServerErr, msg),
	)
}

// readClientMessages reads JSON messages from the WebSocket client
// and dispatches input/resize commands to the SSH session.
func readClientMessages(
	clientConn wsConn, stdinPipe io.WriteCloser, session *ssh.Session,
) {
	for {
		_, raw, err := clientConn.ReadMessage()
		if err != nil {
			return
		}
		var msg terminalMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "input":
			data, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				continue
			}
			if _, err := stdinPipe.Write(data); err != nil {
				return
			}
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				_ = session.WindowChange(int(msg.Rows), int(msg.Cols))
			}
		}
	}
}

// wsConn is the subset of *websocket.Conn used by readClientMessages.
type wsConn interface {
	ReadMessage() (messageType int, p []byte, err error)
}

// loadTerminalConfig reads the terminal widget config and decrypts the
// password when needed. Decryption failures are surfaced explicitly so the
// user is told the encryption key has changed instead of silently sending a
// wrong password to sshd.
func (h *TerminalHandler) loadTerminalConfig(ctx context.Context, widgetID int64) (terminalConfig, error) {
	widget, err := h.DB.GetWidget(ctx, widgetID)
	if err != nil {
		return terminalConfig{}, fmt.Errorf("widget not found: %w", err)
	}

	var cfg terminalConfig
	if err := json.Unmarshal([]byte(widget.Config), &cfg); err != nil {
		return terminalConfig{}, errors.New("invalid widget config")
	}
	if cfg.Host == "" || cfg.Username == "" {
		return terminalConfig{}, errors.New("SSH host and username required")
	}
	if cfg.Port == 0 {
		cfg.Port = 22
	}

	if err := h.maybeDecryptPassword(ctx, widget.Config, &cfg); err != nil {
		return terminalConfig{}, err
	}
	return cfg, nil
}

// maybeDecryptPassword decrypts cfg.Password in place when the widget config
// is marked _encrypted. Any decryption failure (wrong key, corrupted blob,
// missing system setting) is returned with a user-facing message.
func (h *TerminalHandler) maybeDecryptPassword(
	ctx context.Context, rawConfig string, cfg *terminalConfig,
) error {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(rawConfig), &raw); err != nil {
		return nil
	}
	encrypted, _ := raw["_encrypted"].(bool)
	if !encrypted {
		return nil
	}

	encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
	if err != nil || encKey == "" {
		return errors.New("failed to decrypt SSH password: encryption key not available")
	}
	decrypted, err := crypto.Decrypt(cfg.Password, encKey)
	if err != nil {
		return errors.New(
			"failed to decrypt SSH password: encryption key mismatch — re-enter the password in the widget settings",
		)
	}
	cfg.Password = decrypted
	return nil
}

// dialSSH establishes an SSH connection using the widget terminal config.
func dialSSH(cfg terminalConfig) (*ssh.Client, *ssh.Session, error) {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	config := &ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(cfg.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // self-hosted LAN
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

// classifyDialError converts a low-level SSH/network error into a short
// human-readable message displayed in the terminal widget. It distinguishes
// authentication failures from network reachability and timeout problems so
// the user can act on it without digging into server logs.
func classifyDialError(err error, host string, port int) string {
	if err == nil {
		return ""
	}
	target := net.JoinHostPort(host, strconv.Itoa(port))
	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "unable to authenticate"):
		return "SSH authentication failed: invalid username or password for " + target
	case strings.Contains(lower, "connection refused"):
		return "SSH connection refused on " + target + " (sshd not running or port blocked)"
	case strings.Contains(lower, "i/o timeout"),
		strings.Contains(lower, "connection timed out"),
		strings.Contains(lower, "deadline exceeded"):
		return "SSH connection timeout reaching " + target
	case strings.Contains(lower, "no route to host"):
		return "SSH host unreachable: " + target
	case strings.Contains(lower, "no such host"):
		return "SSH host not found: " + host
	case strings.Contains(lower, "handshake failed"):
		return "SSH handshake failed with " + target + ": " + msg
	default:
		return "SSH error: " + msg
	}
}
