package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the configuration directory path following XDG spec.
// Priority: $XDG_CONFIG_HOME/loom > ~/.config/loom > ~/.loom
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "loom")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".loom"
	}

	// Prefer ~/.config/loom if .config exists
	configDir := filepath.Join(home, ".config")
	if info, err := os.Stat(configDir); err == nil && info.IsDir() {
		return filepath.Join(configDir, "loom")
	}

	return filepath.Join(home, ".loom")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// ThemesDir returns the directory for custom themes.
func ThemesDir() string {
	return filepath.Join(ConfigDir(), "themes")
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	dir := ConfigDir()
	return os.MkdirAll(dir, 0755)
}

// EnsureThemesDir creates the themes directory if it doesn't exist.
func EnsureThemesDir() error {
	dir := ThemesDir()
	return os.MkdirAll(dir, 0755)
}
