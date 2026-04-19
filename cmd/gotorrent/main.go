// GoTorrent — a fast, lightweight BitTorrent desktop client.
// Built with Go and the Fyne v2 UI framework.
package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tarunvishwakarma1/gotorrent/internal/config"
	"github.com/tarunvishwakarma1/gotorrent/internal/engine"
	"github.com/tarunvishwakarma1/gotorrent/internal/ipc"
	"github.com/tarunvishwakarma1/gotorrent/internal/ui"
)

func main() {
	// ── 1. Config ──────────────────────────────────────────────────
	cfgMgr, err := config.NewManager()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── 2. Logging ────────────────────────────────────────────────
	setupLogging()

	// ── 3. Single-instance check ──────────────────────────────────
	// torrentArg is the optional .torrent path passed via CLI or file association.
	var torrentArg string
	if len(os.Args) > 1 {
		torrentArg = os.Args[1]
	}

	// ipcServer will be non-nil only for the first instance.
	var ipcServer *ipc.Server
	if ipc.IsRunning() {
		// Another instance is running. Forward our file (if any) and exit.
		if torrentArg != "" {
			if err := ipc.SendPath(torrentArg); err != nil {
				log.Printf("ipc: send path: %v", err)
			}
		}
		os.Exit(0)
	}

	// ── 4. Engine ─────────────────────────────────────────────────
	cfg := cfgMgr.Get()

	statePath, err := config.StatePath()
	if err != nil {
		log.Fatalf("state path: %v", err)
	}

	mgr := engine.NewManager(cfg.MaxConcurrent, statePath, cfg.AutoStart)
	if err := mgr.LoadState(); err != nil {
		log.Printf("load state: %v", err)
	}

	// ── 5. UI ─────────────────────────────────────────────────────
	gta := ui.New(mgr, cfgMgr)

	// ── 6. IPC server (first instance) ───────────────────────────
	ipcServer, err = ipc.TryBecomeServer(func(path string) {
		// Called when another instance sends us a torrent path.
		gta.OpenTorrentFile(path)
	})
	if err != nil {
		log.Printf("ipc listen: %v", err)
	}
	if ipcServer != nil {
		defer ipcServer.Close()
	}

	// ── 7. Run ───────────────────────────────────────────────────
	gta.ShowAndRun(torrentArg)

	// ── 8. Graceful shutdown ──────────────────────────────────────
	mgr.Shutdown()
}

// setupLogging configures the global logger to write to both stderr and a file.
func setupLogging() {
	logPath := logFilePath()
	if logPath == "" {
		return
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("GoTorrent starting")
}

// logFilePath returns the platform-appropriate log file path.
func logFilePath() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "GoTorrent", "gotorrent.log")
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "GoTorrent", "gotorrent.log")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "GoTorrent", "gotorrent.log")
	}
}
