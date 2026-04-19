// Package config manages persistent application settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// ThemeChoice represents the user's theme preference.
type ThemeChoice string

const (
	ThemeLight  ThemeChoice = "light"
	ThemeDark   ThemeChoice = "dark"
	ThemeSystem ThemeChoice = "system"
)

// Config holds all user-configurable settings.
type Config struct {
	SavePath         string      `json:"save_path"`
	MaxConcurrent    int         `json:"max_concurrent"`
	MaxConnections   int         `json:"max_connections"`
	ListenPort       int         `json:"listen_port"`
	Theme            ThemeChoice `json:"theme"`
	StartMinimized   bool        `json:"start_minimized"`
	NotifyOnComplete bool        `json:"notify_on_complete"`
	MinimizeToTray   bool        `json:"minimize_to_tray"`
	AutoStart        bool        `json:"auto_start"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		SavePath:         filepath.Join(home, "Downloads"),
		MaxConcurrent:    3,
		MaxConnections:   50,
		ListenPort:       6881,
		Theme:            ThemeSystem,
		StartMinimized:   false,
		NotifyOnComplete: true,
		MinimizeToTray:   true,
		AutoStart:        true,
	}
}

// Manager persists and provides access to the application Config.
type Manager struct {
	mu   sync.RWMutex
	cfg  *Config
	path string
}

// NewManager creates a Manager, loading config from disk if available.
// The config file lives in the OS-appropriate app data directory.
func NewManager() (*Manager, error) {
	dir, err := AppDataDir()
	if err != nil {
		return nil, fmt.Errorf("config: get app data dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("config: create dir: %w", err)
	}

	m := &Manager{
		path: filepath.Join(dir, "config.json"),
		cfg:  DefaultConfig(),
	}
	_ = m.load() // ignore on first run
	return m, nil
}

// Get returns a copy of the current config.
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c := *m.cfg
	return &c
}

// Save persists a new config to disk.
func (m *Manager) Save(cfg *Config) error {
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return m.write()
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, m.cfg)
}

func (m *Manager) write() error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.cfg, "", "  ")
	m.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	// atomic write: temp file + rename
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("config: write temp: %w", err)
	}
	return os.Rename(tmp, m.path)
}

// AppDataDir returns the OS-appropriate directory for GoTorrent data.
func AppDataDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "GoTorrent"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA env not set")
		}
		return filepath.Join(appData, "GoTorrent"), nil
	default: // linux and others
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		return filepath.Join(home, ".local", "share", "GoTorrent"), nil
	}
}

// LogPath returns the path to the application log file.
func LogPath() (string, error) {
	dir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gotorrent.log"), nil
}

// StatePath returns the path to the torrent state persistence file.
func StatePath() (string, error) {
	dir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "torrents.json"), nil
}
