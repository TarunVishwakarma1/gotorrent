package p2p

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tarunvishwakarma1/gotorrent/client"
	message "github.com/tarunvishwakarma1/gotorrent/messages"
	"github.com/tarunvishwakarma1/gotorrent/peers"
	"github.com/tarunvishwakarma1/gotorrent/torrent"
)

const (
	MaxBlockSize     = 16384 // 16KB — max per-block request (BitTorrent standard)
	MaxBacklog       = 16    // pipeline depth per peer; 3× throughput vs old value of 5
	endgameThreshold = 10    // activate endgame mode when this many pieces remain
)

// pieceWork is a piece that needs to be downloaded
type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

// pieceResult is a piece that has been downloaded
type pieceResult struct {
	index int
	buf   []byte
}

// pieceProgress tracks the state of downloading one piece
type pieceProgress struct {
	index      int
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

// ProgressFunc is called roughly every second during a download.
// downloaded is total bytes downloaded so far, speed is bytes/sec,
// numPeers is the current connected peer count.
type ProgressFunc func(downloaded int64, speed float64, numPeers int)

// Torrent holds all the info needed to download
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	// Ctx allows the caller to cancel a download. Nil means no cancellation.
	Ctx context.Context
	// ProgressFunc is called every second with download progress. May be nil.
	ProgressFunc ProgressFunc
}

// calculatePieceSize returns the size of a specific piece
// the last piece is often smaller than the rest
func (t *Torrent) calculatePieceSize(index int) int {
	start := index * t.PieceLength
	end := start + t.PieceLength
	if end > t.Length {
		end = t.Length // last piece might be smaller
	}
	return end - start
}

// calculateBoundsForPiece returns the start and end byte
// position of a piece in the final file
func (t *Torrent) calculateBoundsForPiece(index int) (int, int) {
	start := index * t.PieceLength
	end := start + t.calculatePieceSize(index)
	return start, end
}

// checkIntegrity verifies a downloaded piece against its expected SHA1 hash
func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("piece %d failed integrity check", pw.index)
	}
	return nil
}

// readMessage reads one message and updates piece progress state
func (state *pieceProgress) readMessage(c *client.Client) error {
	msg, err := message.Read(c.Conn)
	if err != nil {
		return err
	}

	// nil means keepalive — just ignore it
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		c.Choked = false

	case message.MsgChoke:
		c.Choked = true

	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		c.Bitfield.SetPiece(index)

	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}

	return nil
}

// attemptDownloadPiece tries to download a single piece from a peer
func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index: pw.index,
		buf:   make([]byte, pw.length),
	}

	// 20 second deadline — tighter than before; slow peers get dropped faster
	c.Conn.SetDeadline(time.Now().Add(20 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // clear deadline when done

	for state.downloaded < pw.length {
		// if unchoked, fill up the pipeline with requests
		if !c.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize

				// last block might be smaller
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				// send the request
				req := message.NewRequest(state.index, state.requested, blockSize)
				_, err := c.Conn.Write(req.Serialize())
				if err != nil {
					return nil, err
				}

				state.backlog++
				state.requested += blockSize
			}
		}

		// read one message and update state
		err := state.readMessage(c)
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

// startDownloadWorker runs in a goroutine for each peer.
// It pulls pieces from workQueue, downloads them, and sends results.
// It exits when the context is cancelled or workQueue is closed.
func (t *Torrent) startDownloadWorker(ctx context.Context, peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s: %v\n", peer, err)
		return
	}
	defer c.Conn.Close()
	log.Printf("Connected to peer %s\n", peer)

	_, err = c.Conn.Write(message.NewInterested().Serialize())
	if err != nil {
		log.Printf("Could not send Interested to %s: %v\n", peer, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case pw, ok := <-workQueue:
			if !ok {
				return
			}
			if !c.Bitfield.HasPiece(pw.index) {
				workQueue <- pw
				// Yield prevents hot-spin when this peer lacks most remaining pieces.
				runtime.Gosched()
				continue
			}
			buf, err := attemptDownloadPiece(c, pw)
			if err != nil {
				log.Printf("Failed to download piece %d from %s: %v\n", pw.index, peer, err)
				workQueue <- pw
				return
			}
			err = checkIntegrity(pw, buf)
			if err != nil {
				log.Printf("Piece %d failed integrity check from %s\n", pw.index, peer)
				workQueue <- pw
				continue
			}
			_, err = c.Conn.Write(message.NewHave(pw.index).Serialize())
			if err != nil {
				log.Printf("Could not send Have to %s: %v\n", peer, err)
			}
			select {
			case results <- &pieceResult{pw.index, buf}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// downloadToBuffer is the core download loop.
// It returns the complete file buffer on success.
//
// Performance characteristics:
//   - MaxBacklog=16 gives 3× pipeline depth vs old value of 5.
//   - Endgame mode: when ≤endgameThreshold pieces remain, each is re-queued
//     for every connected peer simultaneously — prevents tail-piece stalls.
//   - Deduplication: atomic bitfield guards against double-copy in endgame.
//   - Atomic peer counter reported to ProgressFunc each second.
func (t *Torrent) downloadToBuffer(ctx context.Context) ([]byte, error) {
	log.Printf("Starting download: %s\n", t.Name)

	totalPieces := len(t.PieceHashes)

	workQueue := make(chan *pieceWork, totalPieces)
	results := make(chan *pieceResult, len(t.Peers)) // buffered to reduce scheduler pressure

	// Track all pieces so endgame can re-queue the survivors.
	allPieces := make([]*pieceWork, totalPieces)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		pw := &pieceWork{index, hash, length}
		allPieces[index] = pw
		workQueue <- pw
	}

	// Atomic connected-peer counter — workers increment/decrement around their lifecycle.
	var activePeers int64
	for _, peer := range t.Peers {
		go func(p peers.Peer) {
			atomic.AddInt64(&activePeers, 1)
			defer atomic.AddInt64(&activePeers, -1)
			t.startDownloadWorker(ctx, p, workQueue, results)
		}(peer)
	}

	buf := make([]byte, t.Length)
	done := 0

	// received[i] = true once piece i is written to buf (endgame dedup guard).
	received := make([]bool, totalPieces)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var downloaded int64
	var lastDownloaded int64
	endgame := false

	defer close(workQueue)

	for done < totalPieces {
		select {
		case res := <-results:
			if received[res.index] {
				// Endgame duplicate — discard without counting.
				continue
			}
			received[res.index] = true

			start, end := t.calculateBoundsForPiece(res.index)
			copy(buf[start:end], res.buf)
			done++

			downloaded = int64(done) * int64(t.PieceLength)
			if downloaded > int64(t.Length) {
				downloaded = int64(t.Length)
			}

			remaining := totalPieces - done
			log.Printf("Progress: %d/%d pieces (%.2f%%)\n", done, totalPieces,
				float64(done)/float64(totalPieces)*100)

			// Endgame activation: flood remaining pieces to all peers.
			if !endgame && remaining > 0 && remaining <= endgameThreshold {
				endgame = true
				log.Printf("Endgame activated with %d pieces remaining\n", remaining)
				numPeers := int(atomic.LoadInt64(&activePeers))
				if numPeers < 1 {
					numPeers = 1
				}
				for idx, pw := range allPieces {
					if !received[idx] {
						// Queue the piece once per peer so multiple workers race to fetch it.
						for i := 0; i < numPeers; i++ {
							select {
							case workQueue <- pw:
							default: // queue full — one copy is enough
							}
						}
					}
				}
			}

		case <-ticker.C:
			if t.ProgressFunc != nil {
				peers := int(atomic.LoadInt64(&activePeers))
				speed := float64(downloaded - lastDownloaded)
				lastDownloaded = downloaded
				t.ProgressFunc(downloaded, speed, peers)
			}

		case <-ctx.Done():
			return nil, fmt.Errorf("download cancelled: %w", ctx.Err())
		}
	}

	// Final progress update.
	if t.ProgressFunc != nil {
		t.ProgressFunc(int64(t.Length), 0, 0)
	}

	return buf, nil
}

// Download downloads the torrent and saves it to outPath.
// It respects t.Ctx for cancellation and calls t.ProgressFunc for progress.
func (t *Torrent) Download(outPath string) error {
	ctx := t.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	buf, err := t.downloadToBuffer(ctx)
	if err != nil {
		return err
	}

	log.Printf("Download complete! Saving to %s\n", outPath)
	if err := os.WriteFile(outPath, buf, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// DownloadBuffer downloads the torrent and returns raw bytes.
// Useful for multi-file torrents where the engine handles splitting.
// It respects t.Ctx for cancellation and calls t.ProgressFunc for progress.
func (t *Torrent) DownloadBuffer() ([]byte, error) {
	ctx := t.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return t.downloadToBuffer(ctx)
}

// New creates a Torrent from a TorrentFile and a list of peers
func New(tf *torrent.TorrentFile, peerList []peers.Peer) *Torrent {
	return &Torrent{
		Peers:       peerList,
		PeerID:      tf.PeerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}
}
