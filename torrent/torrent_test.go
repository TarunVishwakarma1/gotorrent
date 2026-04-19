package torrent

import (
	"crypto/sha1"
	"strings"
	"testing"

	"github.com/tarunvishwakarma1/gotorret/parser"
)

// buildTorrent constructs a valid bencoded torrent string for testing.
// pieces must be a string whose length is a multiple of 20.
func buildTorrent(announce, name string, length, pieceLength int, pieces string) string {
	info := map[string]any{
		"length":       length,
		"name":         name,
		"piece length": pieceLength,
		"pieces":       pieces,
	}
	m := map[string]any{
		"announce": announce,
		"info":     info,
	}
	return parser.Encode(m)
}

// onePiece is a valid 20-byte piece hash string.
const onePiece = "AAAAAAAAAAAAAAAAAAAA"

func TestNewTorrentFileValid(t *testing.T) {
	tstr := buildTorrent("http://tracker.example.com/announce", "test.iso", 1000, 256, onePiece)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tf.Announce != "http://tracker.example.com/announce" {
		t.Errorf("announce: expected %q, got %q", "http://tracker.example.com/announce", tf.Announce)
	}
	if tf.Name != "test.iso" {
		t.Errorf("name: expected %q, got %q", "test.iso", tf.Name)
	}
	if tf.Length != 1000 {
		t.Errorf("length: expected 1000, got %d", tf.Length)
	}
	if tf.PieceLength != 256 {
		t.Errorf("piece length: expected 256, got %d", tf.PieceLength)
	}
}

func TestNewTorrentFilePieceHashes(t *testing.T) {
	// Two pieces = 40 bytes
	twoPieces := onePiece + "BBBBBBBBBBBBBBBBBBBB"
	tstr := buildTorrent("http://tracker.example.com/announce", "test.iso", 1000, 256, twoPieces)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tf.PieceHashes) != 2 {
		t.Fatalf("expected 2 piece hashes, got %d", len(tf.PieceHashes))
	}
	// Verify first hash bytes
	for i, b := range tf.PieceHashes[0] {
		if b != 'A' {
			t.Errorf("piece[0][%d]: expected 'A', got %q", i, b)
		}
	}
	// Verify second hash bytes
	for i, b := range tf.PieceHashes[1] {
		if b != 'B' {
			t.Errorf("piece[1][%d]: expected 'B', got %q", i, b)
		}
	}
}

func TestNewTorrentFileSinglePieceHash(t *testing.T) {
	tstr := buildTorrent("http://tracker.example.com/announce", "file.bin", 512, 512, onePiece)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tf.PieceHashes) != 1 {
		t.Errorf("expected 1 piece hash, got %d", len(tf.PieceHashes))
	}
}

func TestNewTorrentFileInfoHashDeterministic(t *testing.T) {
	tstr := buildTorrent("http://tracker.example.com/announce", "test.iso", 1000, 256, onePiece)
	tf1, _ := NewTorrentFile(tstr)
	tf2, _ := NewTorrentFile(tstr)
	if tf1.InfoHash != tf2.InfoHash {
		t.Errorf("info hash is not deterministic: %x vs %x", tf1.InfoHash, tf2.InfoHash)
	}
}

func TestNewTorrentFileInfoHashValue(t *testing.T) {
	// Compute expected info hash manually
	info := map[string]any{
		"length":       1000,
		"name":         "test.iso",
		"piece length": 256,
		"pieces":       onePiece,
	}
	rawInfo := parser.Encode(info)
	expectedHash := sha1.Sum([]byte(rawInfo))

	tstr := buildTorrent("http://tracker.example.com/announce", "test.iso", 1000, 256, onePiece)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tf.InfoHash != expectedHash {
		t.Errorf("info hash mismatch: expected %x, got %x", expectedHash, tf.InfoHash)
	}
}

func TestNewTorrentFileMissingInfoDict(t *testing.T) {
	// Bencode without "info" key
	encoded := parser.Encode(map[string]any{
		"announce": "http://tracker.example.com/announce",
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for missing info dict, got nil")
	}
}

func TestNewTorrentFileMissingAnnounce(t *testing.T) {
	encoded := parser.Encode(map[string]any{
		"info": map[string]any{
			"length":       1000,
			"name":         "test.iso",
			"piece length": 256,
			"pieces":       onePiece,
		},
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for missing announce, got nil")
	}
}

func TestNewTorrentFileMissingName(t *testing.T) {
	encoded := parser.Encode(map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"length":       1000,
			"piece length": 256,
			"pieces":       onePiece,
		},
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestNewTorrentFileMissingLength(t *testing.T) {
	encoded := parser.Encode(map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.iso",
			"piece length": 256,
			"pieces":       onePiece,
		},
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for missing length, got nil")
	}
}

func TestNewTorrentFileMissingPieceLength(t *testing.T) {
	encoded := parser.Encode(map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"length":  1000,
			"name":    "test.iso",
			"pieces":  onePiece,
		},
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for missing piece length, got nil")
	}
}

func TestNewTorrentFileMissingPieces(t *testing.T) {
	encoded := parser.Encode(map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"length":       1000,
			"name":         "test.iso",
			"piece length": 256,
		},
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for missing pieces, got nil")
	}
}

func TestNewTorrentFilePiecesNotMultipleOf20(t *testing.T) {
	encoded := parser.Encode(map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"length":       1000,
			"name":         "test.iso",
			"piece length": 256,
			"pieces":       "AAAAAAAAAA", // 10 bytes, not multiple of 20
		},
	})
	_, err := NewTorrentFile(encoded)
	if err == nil {
		t.Error("expected error for pieces length not multiple of 20, got nil")
	}
}

func TestNewTorrentFileMultiplePieces(t *testing.T) {
	// 5 pieces = 100 bytes
	pieces := strings.Repeat("A", 100)
	tstr := buildTorrent("http://tracker.example.com/announce", "big.iso", 5000, 1000, pieces)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tf.PieceHashes) != 5 {
		t.Errorf("expected 5 piece hashes, got %d", len(tf.PieceHashes))
	}
}

func TestNewTorrentFilePiecesExactly20(t *testing.T) {
	// Boundary: exactly 20 bytes → exactly 1 piece hash, no error
	tstr := buildTorrent("http://tracker.example.com/announce", "exact.iso", 100, 100, onePiece)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error for exactly 20-byte pieces: %v", err)
	}
	if len(tf.PieceHashes) != 1 {
		t.Errorf("expected 1 piece hash, got %d", len(tf.PieceHashes))
	}
}