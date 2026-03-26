package config

import (
	"os"
	"path/filepath"
)

// DesktopDataDir returns the OS-specific data directory for desktop mode.
func DesktopDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "nidus"), nil
}
