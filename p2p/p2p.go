package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tarunvishwakarma1/gotorret/client"
	message "github.com/tarunvishwakarma1/gotorret/messages"
	"github.com/tarunvishwakarma1/gotorret/peers"
	"github.com/tarunvishwakarma1/gotorret/torrent"
)

const (
	MaxBlockSize = 16384 // 16KB — max bytes per request
	MaxBacklog   = 5     // max number of unfulfilled requests
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

// Torrent holds all the info needed to download
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
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

	// set a 30 second deadline — if a piece takes longer something is wrong
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // disable deadline when done

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

// startDownloadWorker runs in a goroutine for each peer
// it pulls pieces from workQueue, downloads them, and sends results
func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	// connect and handshake
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s: %v\n", peer, err)
		return
	}
	defer c.Conn.Close()
	log.Printf("Connected to peer %s\n", peer)

	// tell the peer we are interested
	_, err = c.Conn.Write(message.NewInterested().Serialize())
	if err != nil {
		log.Printf("Could not send Interested to %s: %v\n", peer, err)
		return
	}

	// pull pieces from the work queue
	for pw := range workQueue {
		// skip if this peer doesn't have the piece
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw // put it back for another peer
			continue
		}

		// try to download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Printf("Failed to download piece %d from %s: %v\n", pw.index, peer, err)
			workQueue <- pw // put it back for another peer
			return
		}

		// verify the hash
		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece %d failed integrity check from %s\n", pw.index, peer)
			workQueue <- pw // corrupted — try again with another peer
			continue
		}

		// tell the peer we got the piece
		_, err = c.Conn.Write(message.NewHave(pw.index).Serialize())
		if err != nil {
			log.Printf("Could not send Have to %s: %v\n", peer, err)
		}

		// send result to main goroutine
		results <- &pieceResult{pw.index, buf}
	}
}

// Download downloads the torrent and saves it to a file
func (t *Torrent) Download(outPath string) error {
	log.Printf("Starting download: %s\n", t.Name)

	// fill the work queue with all pieces
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// spawn one goroutine per peer
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// collect results into a buffer
	buf := make([]byte, t.Length)
	donePieces := 0
	totalPieces := len(t.PieceHashes)

	for donePieces < totalPieces {
		res := <-results
		start, end := t.calculateBoundsForPiece(res.index)
		copy(buf[start:end], res.buf)
		donePieces++

		// print progress
		percent := float64(donePieces) / float64(totalPieces) * 100
		log.Printf("Progress: %d/%d pieces (%.2f%%)\n", donePieces, totalPieces, percent)
	}
	close(workQueue)

	// write the buffer to a file
	log.Printf("Download complete! Saving to %s\n", outPath)
	err := os.WriteFile(outPath, buf, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
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
