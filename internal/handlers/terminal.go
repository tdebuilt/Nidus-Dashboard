package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"

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
func (h *TerminalHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.getWidgetConfig(r)
	if err != nil {
		slog.Error("terminal: config error", "error", err)
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	sshClient, session, err := dialSSH(cfg)
	if err != nil {
		slog.Error("terminal: SSH dial failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "SSH connection failed"})
		return
	}
	defer sshClient.Close()
	defer session.Close()

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "SSH session error"})
		return
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "SSH session error"})
		return
	}

	if err := session.RequestPty("xterm-256color", 24, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "PTY request failed"})
		return
	}

	if err := session.Shell(); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "shell start failed"})
		return
	}

	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("terminal: WebSocket upgrade failed", "error", err)
		return
	}
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// SSH stdout -> WebSocket (binary frames)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, readErr := stdoutPipe.Read(buf)
			if n > 0 {
				if writeErr := clientConn.WriteMessage(2, buf[:n]); writeErr != nil {
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	// WebSocket -> SSH stdin (JSON messages)
	readClientMessages(clientConn, stdinPipe, session)

	session.Close()
	wg.Wait()
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

// getWidgetConfig reads the terminal widget config from the database.
func (h *TerminalHandler) getWidgetConfig(r *http.Request) (terminalConfig, error) {
	widgetIDStr := r.URL.Query().Get("widget_id")
	if widgetIDStr == "" {
		return terminalConfig{}, errors.New("widget_id required")
	}
	widgetID, err := strconv.ParseInt(widgetIDStr, 10, 64)
	if err != nil {
		return terminalConfig{}, errors.New("invalid widget_id")
	}

	widget, err := h.DB.GetWidget(r.Context(), widgetID)
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

	// Decrypt password if encrypted
	var raw map[string]interface{}
	if json.Unmarshal([]byte(widget.Config), &raw) == nil {
		if encrypted, ok := raw["_encrypted"].(bool); ok && encrypted {
			encKey, err := h.DB.GetSystemSetting(r.Context(), "encryption_key")
			if err == nil && encKey != "" {
				if decrypted, err := crypto.Decrypt(cfg.Password, encKey); err == nil {
					cfg.Password = decrypted
				}
			}
		}
	}

	return cfg, nil
}

// dialSSH establishes an SSH connection using the widget terminal config.
func dialSSH(cfg terminalConfig) (*ssh.Client, *ssh.Session, error) {
	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))

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
