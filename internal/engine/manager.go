package engine

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tarunvishwakarma1/gotorrent/torrent"
)

// TorrentManager manages the lifecycle of all torrents: adding, pausing,
// resuming, removing, and persisting state across restarts.
type TorrentManager struct {
	mu        sync.RWMutex
	torrents  map[string]*TorrentState
	cancels   map[string]context.CancelFunc
	sem       chan struct{} // bounds max concurrent downloads
	statePath string
	autoStart bool

	// OnUpdate is called (possibly from a goroutine) whenever a torrent's
	// state changes. The caller must not hold any lock when invoking it.
	OnUpdate func(state *TorrentState)
	// OnComplete is called when a torrent finishes successfully.
	OnComplete func(state *TorrentState)
	// OnError is called when a download fails.
	OnError func(state *TorrentState)
}

// NewManager creates a TorrentManager.
// maxConcurrent limits simultaneous active downloads.
// statePath is the JSON file used for state persistence.
// autoStart controls whether newly added torrents begin immediately.
func NewManager(maxConcurrent int, statePath string, autoStart bool) *TorrentManager {
	return &TorrentManager{
		torrents:  make(map[string]*TorrentState),
		cancels:   make(map[string]context.CancelFunc),
		sem:       make(chan struct{}, maxConcurrent),
		statePath: statePath,
		autoStart: autoStart,
	}
}

// LoadState restores previously saved torrent states from disk.
// Torrents that were downloading are set to Paused (user must resume).
func (m *TorrentManager) LoadState() error {
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load state: %w", err)
	}

	var states []*TorrentState
	if err := json.Unmarshal(data, &states); err != nil {
		return fmt.Errorf("parse state: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range states {
		if s.Status == StatusDownloading || s.Status == StatusConnecting {
			s.Status = StatusPaused
		}
		m.torrents[s.ID] = s
	}
	return nil
}

// Add parses a .torrent file and registers it with the manager.
// If the torrent is already tracked (same InfoHash) it returns an error.
// When autoStart is true the download begins immediately.
func (m *TorrentManager) Add(torrentPath, savePath string) (*TorrentState, error) {
	data, err := os.ReadFile(torrentPath)
	if err != nil {
		return nil, fmt.Errorf("read torrent file: %w", err)
	}

	tf, err := torrent.NewTorrentFile(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse torrent: %w", err)
	}

	infoHashHex := hex.EncodeToString(tf.InfoHash[:])

	m.mu.Lock()
	for _, s := range m.torrents {
		if s.InfoHash == infoHashHex {
			m.mu.Unlock()
			return nil, fmt.Errorf("torrent already exists: %s", tf.Name)
		}
	}

	files := make([]FileState, 0, len(tf.Files))
	for _, f := range tf.Files {
		name := ""
		if len(f.Path) > 0 {
			name = f.Path[len(f.Path)-1]
		}
		files = append(files, FileState{
			Path:     f.Path,
			Length:   f.Length,
			Selected: true,
			Icon:     FileIconForName(name),
		})
	}

	state := &TorrentState{
		ID:          uuid.NewString(),
		Name:        tf.Name,
		TorrentPath: torrentPath,
		SavePath:    savePath,
		TotalSize:   int64(tf.Length),
		Status:      StatusQueued,
		AddedAt:     time.Now(),
		InfoHash:    infoHashHex,
		IsMultiFile: tf.IsMultiFile,
		Files:       files,
		ETA:         "Unknown",
	}

	m.torrents[state.ID] = state
	m.mu.Unlock()

	m.persist()
	m.notify(state)

	if m.autoStart {
		go m.startDownload(state.ID)
	}

	return state.Clone(), nil
}

// Pause cancels an active download and marks the torrent as Paused.
func (m *TorrentManager) Pause(id string) error {
	m.mu.Lock()
	state, ok := m.torrents[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("torrent not found: %s", id)
	}
	if state.Status != StatusDownloading && state.Status != StatusConnecting {
		m.mu.Unlock()
		return fmt.Errorf("torrent not active: %s", state.Status)
	}
	cancel, hasCanel := m.cancels[id]
	m.mu.Unlock()

	if hasCanel {
		cancel()
	}
	return nil
}

// Resume restarts a paused torrent download.
func (m *TorrentManager) Resume(id string) error {
	m.mu.Lock()
	state, ok := m.torrents[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("torrent not found: %s", id)
	}
	if state.Status != StatusPaused && state.Status != StatusQueued && state.Status != StatusError {
		m.mu.Unlock()
		return fmt.Errorf("torrent not paused: %s", state.Status)
	}
	m.mu.Unlock()

	go m.startDownload(id)
	return nil
}

// Remove stops and removes a torrent from the manager.
// If deleteFiles is true, the downloaded output is deleted from disk.
func (m *TorrentManager) Remove(id string, deleteFiles bool) error {
	m.mu.Lock()
	state, ok := m.torrents[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("torrent not found: %s", id)
	}

	cancel, hasCancel := m.cancels[id]
	delete(m.torrents, id)
	delete(m.cancels, id)
	savePath := state.SavePath
	name := state.Name
	m.mu.Unlock()

	if hasCancel {
		cancel()
	}

	if deleteFiles {
		target := fmt.Sprintf("%s/%s", savePath, name)
		if err := os.RemoveAll(target); err != nil {
			log.Printf("engine: remove files %s: %v", target, err)
		}
	}

	m.persist()
	return nil
}

// GetAll returns clones of all tracked torrent states.
func (m *TorrentManager) GetAll() []*TorrentState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*TorrentState, 0, len(m.torrents))
	for _, s := range m.torrents {
		out = append(out, s.Clone())
	}
	return out
}

// Get returns a clone of a single torrent state.
func (m *TorrentManager) Get(id string) (*TorrentState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.torrents[id]
	if !ok {
		return nil, fmt.Errorf("torrent not found: %s", id)
	}
	return s.Clone(), nil
}

// SetMaxConcurrent adjusts the concurrency limit.
// Takes effect for future downloads; does not stop current ones.
func (m *TorrentManager) SetMaxConcurrent(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sem = make(chan struct{}, n)
}

// Shutdown gracefully cancels all active downloads.
func (m *TorrentManager) Shutdown() {
	m.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(m.cancels))
	for _, cancel := range m.cancels {
		cancels = append(cancels, cancel)
	}
	m.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
	m.persist()
}

// startDownload runs the download pipeline for torrent id in a goroutine.
// It acquires the semaphore, updates state transitions, and handles errors.
func (m *TorrentManager) startDownload(id string) {
	// Acquire concurrency slot.
	m.sem <- struct{}{}
	defer func() { <-m.sem }()

	m.setStatus(id, StatusConnecting, "")

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[id] = cancel
	state, ok := m.torrents[id]
	if !ok {
		m.mu.Unlock()
		cancel()
		return
	}
	stateCopy := state.Clone()
	m.mu.Unlock()

	err := RunDownload(ctx, stateCopy, func(downloaded int64, speed float64, numPeers int) {
		m.mu.Lock()
		s, ok := m.torrents[id]
		if ok {
			s.Downloaded = downloaded
			s.Speed = speed
			s.Peers = numPeers
			s.Status = StatusDownloading
			if s.TotalSize > 0 {
				s.Progress = float64(downloaded) / float64(s.TotalSize)
			}
			remaining := s.TotalSize - downloaded
			if speed > 0 && remaining > 0 {
				s.ETA = FormatETA(float64(remaining) / speed)
			} else {
				s.ETA = "Unknown"
			}
			clone := s.Clone()
			m.mu.Unlock()
			m.notify(clone)
		} else {
			m.mu.Unlock()
		}
	})

	m.mu.Lock()
	delete(m.cancels, id)
	m.mu.Unlock()
	cancel()

	if err != nil {
		if ctx.Err() != nil {
			// Cancelled by Pause or Remove — set Paused.
			m.setStatus(id, StatusPaused, "")
		} else {
			m.setStatus(id, StatusError, err.Error())
			m.mu.RLock()
			s, ok := m.torrents[id]
			var clone *TorrentState
			if ok {
				clone = s.Clone()
			}
			m.mu.RUnlock()
			if clone != nil && m.OnError != nil {
				m.OnError(clone)
			}
		}
	} else {
		m.mu.Lock()
		s, ok := m.torrents[id]
		if ok {
			s.Status = StatusComplete
			s.Progress = 1.0
			s.Downloaded = s.TotalSize
			s.Speed = 0
			s.Peers = 0
			s.ETA = ""
			s.CompletedAt = time.Now()
			clone := s.Clone()
			m.mu.Unlock()
			m.notify(clone)
			if m.OnComplete != nil {
				m.OnComplete(clone)
			}
		} else {
			m.mu.Unlock()
		}
	}

	m.persist()
}

// setStatus updates the status (and optional error string) of a torrent.
func (m *TorrentManager) setStatus(id string, status Status, errMsg string) {
	m.mu.Lock()
	s, ok := m.torrents[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	s.Status = status
	s.Error = errMsg
	clone := s.Clone()
	m.mu.Unlock()

	m.notify(clone)
	m.persist()
}

// notify calls OnUpdate safely outside the manager lock.
func (m *TorrentManager) notify(state *TorrentState) {
	if m.OnUpdate != nil {
		m.OnUpdate(state)
	}
}

// persist writes all torrent states to the JSON state file atomically.
func (m *TorrentManager) persist() {
	m.mu.RLock()
	states := make([]*TorrentState, 0, len(m.torrents))
	for _, s := range m.torrents {
		states = append(states, s.Clone())
	}
	m.mu.RUnlock()

	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		log.Printf("engine: marshal state: %v", err)
		return
	}
	tmp := m.statePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		log.Printf("engine: write state tmp: %v", err)
		return
	}
	if err := os.Rename(tmp, m.statePath); err != nil {
		log.Printf("engine: rename state: %v", err)
	}
}
