// Package engine implements the torrent download manager.
package engine

import (
	"fmt"
	"time"
)

// Status is the lifecycle state of a managed torrent.
type Status string

const (
	StatusQueued      Status = "Queued"
	StatusConnecting  Status = "Connecting"
	StatusDownloading Status = "Downloading"
	StatusVerifying   Status = "Verifying"
	StatusComplete    Status = "Complete"
	StatusPaused      Status = "Paused"
	StatusError       Status = "Error"
)

// FileState holds metadata for a single file within a multi-file torrent.
type FileState struct {
	Path     []string `json:"path"`
	Length   int64    `json:"length"`
	Selected bool     `json:"selected"`
	// Icon is a display emoji chosen by file extension.
	Icon string `json:"icon"`
}

// TorrentState is the complete, serialisable state of one managed torrent.
type TorrentState struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	TorrentPath string      `json:"torrent_path"`
	SavePath    string      `json:"save_path"`
	TotalSize   int64       `json:"total_size"`
	Downloaded  int64       `json:"downloaded"`
	Status      Status      `json:"status"`
	Progress    float64     `json:"progress"` // 0.0–1.0
	Speed       float64     `json:"speed"`    // bytes/sec
	Peers       int         `json:"peers"`
	ETA         string      `json:"eta"`
	InfoHash    string      `json:"info_hash"`
	Error       string      `json:"error,omitempty"`
	AddedAt     time.Time   `json:"added_at"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
	Files       []FileState `json:"files"`
	IsMultiFile bool        `json:"is_multi_file"`
}

// Clone returns a deep copy safe to read without holding the manager lock.
func (s *TorrentState) Clone() *TorrentState {
	c := *s
	c.Files = make([]FileState, len(s.Files))
	copy(c.Files, s.Files)
	return &c
}

// FileIconForName returns a display emoji for a given filename.
func FileIconForName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			ext := name[i:]
			switch ext {
			case ".iso", ".img", ".dmg":
				return "💿"
			case ".mp4", ".mkv", ".avi", ".mov":
				return "🎬"
			case ".mp3", ".flac", ".wav", ".ogg":
				return "🎵"
			case ".pdf":
				return "📕"
			case ".zip", ".rar", ".7z", ".tar", ".gz":
				return "📦"
			case ".jpg", ".jpeg", ".png", ".gif", ".webp":
				return "🖼"
			case ".txt", ".md", ".nfo", ".log":
				return "📄"
			}
			break
		}
	}
	return "📄"
}

// FormatETA formats seconds into a human-readable ETA string.
func FormatETA(seconds float64) string {
	if seconds <= 0 || seconds > 86400*7 {
		return "Unknown"
	}
	s := int(seconds)
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	m := s / 60
	s = s % 60
	if m < 60 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := m / 60
	m = m % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
