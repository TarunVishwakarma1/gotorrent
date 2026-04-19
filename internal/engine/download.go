package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tarunvishwakarma1/gotorrent/p2p"
	"github.com/tarunvishwakarma1/gotorrent/peers"
	"github.com/tarunvishwakarma1/gotorrent/torrent"
	"github.com/tarunvishwakarma1/gotorrent/tracker"
)

// ProgressCallback is called every second during a download with current stats.
type ProgressCallback func(downloaded int64, speed float64, numPeers int)

// RunDownload executes the full download pipeline for the given state.
// It contacts the tracker, fetches peers, and drives p2p.Torrent.
// The context allows the caller (manager) to cancel or pause.
func RunDownload(ctx context.Context, state *TorrentState, cb ProgressCallback) error {
	// 1. Read the .torrent file from disk.
	data, err := os.ReadFile(state.TorrentPath)
	if err != nil {
		return fmt.Errorf("read torrent file: %w", err)
	}

	// 2. Parse the torrent metadata.
	tf, err := torrent.NewTorrentFile(string(data))
	if err != nil {
		return fmt.Errorf("parse torrent: %w", err)
	}

	// 3. Fetch peers from the tracker.
	rawPeers, err := tracker.GetPeers(tf)
	if err != nil {
		return fmt.Errorf("tracker: %w", err)
	}

	// 4. Decode compact peer list.
	peerList, err := peers.Decode(rawPeers)
	if err != nil {
		return fmt.Errorf("decode peers: %w", err)
	}
	if len(peerList) == 0 {
		return fmt.Errorf("no peers found")
	}

	// 5. Build the p2p.Torrent with progress and cancellation support.
	t := p2p.New(tf, peerList)
	t.Ctx = ctx
	t.ProgressFunc = func(downloaded int64, speed float64, numPeers int) {
		if cb != nil {
			cb(downloaded, speed, numPeers)
		}
	}

	// 6. Execute download.
	if tf.IsMultiFile {
		return downloadMultiFile(ctx, t, tf, state.SavePath)
	}
	outPath := filepath.Join(state.SavePath, tf.Name)
	if err := os.MkdirAll(state.SavePath, 0o755); err != nil {
		return fmt.Errorf("create save dir: %w", err)
	}
	return t.Download(outPath)
}

// downloadMultiFile downloads a multi-file torrent and writes each file to disk.
// Files are placed under savePath/torrentName/.
func downloadMultiFile(ctx context.Context, t *p2p.Torrent, tf *torrent.TorrentFile, savePath string) error {
	_ = ctx // context is already propagated into t.Ctx

	buf, err := t.DownloadBuffer()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(savePath, tf.Name)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("create torrent dir: %w", err)
	}

	offset := int64(0)
	for _, fi := range tf.Files {
		if len(fi.Path) == 0 {
			offset += fi.Length
			continue
		}

		// Build the destination path from the path components.
		parts := make([]string, 0, len(fi.Path)+1)
		parts = append(parts, baseDir)
		parts = append(parts, fi.Path...)
		dest := filepath.Join(parts...)

		// Sanitize: don't allow path traversal.
		if !strings.HasPrefix(dest, baseDir) {
			offset += fi.Length
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("create file dir %s: %w", filepath.Dir(dest), err)
		}

		end := offset + fi.Length
		if end > int64(len(buf)) {
			end = int64(len(buf))
		}
		if err := os.WriteFile(dest, buf[offset:end], 0o644); err != nil {
			return fmt.Errorf("write file %s: %w", dest, err)
		}
		offset = end
	}

	return nil
}
