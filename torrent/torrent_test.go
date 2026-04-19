package torrent

import (
	"crypto/sha1"
	"strings"
	"testing"

	"github.com/tarunvishwakarma1/gotorret/parser"
)

// buildTorrentBencode constructs a minimal valid bencode torrent string.
// pieces must be a string whose length is a multiple of 20.
func buildTorrentBencode(announce, name string, length, pieceLength int, pieces string) string {
	// Manually craft bencode to avoid dependency on Encode for test setup
	// We use Encode here since it's already tested and is the only way to
	// produce the canonical info hash.
	info := map[string]any{
		"length":       length,
		"name":         name,
		"piece length": pieceLength,
		"pieces":       pieces,
	}
	top := map[string]any{
		"announce": announce,
		"info":     info,
	}
	return parser.Encode(top)
}

// pieces20 returns a string of n*20 repeated 'A' bytes.
func pieces20(n int) string {
	return strings.Repeat("AAAAAAAAAAAAAAAAAAAA", n) // each repetition is 20 chars
}

func TestNewTorrentFile_Valid(t *testing.T) {
	announce := "http://tracker.example.com/announce"
	name := "test.iso"
	length := 1000
	pieceLength := 262144
	pieces := pieces20(1)

	input := buildTorrentBencode(announce, name, length, pieceLength, pieces)

	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf.Announce != announce {
		t.Errorf("Announce: expected %q, got %q", announce, tf.Announce)
	}
	if tf.Name != name {
		t.Errorf("Name: expected %q, got %q", name, tf.Name)
	}
	if tf.Length != length {
		t.Errorf("Length: expected %d, got %d", length, tf.Length)
	}
	if tf.PieceLength != pieceLength {
		t.Errorf("PieceLength: expected %d, got %d", pieceLength, tf.PieceLength)
	}
	if len(tf.PieceHashes) != 1 {
		t.Errorf("PieceHashes: expected 1, got %d", len(tf.PieceHashes))
	}
}

func TestNewTorrentFile_MultiplePieces(t *testing.T) {
	pieces := pieces20(3)
	input := buildTorrentBencode("http://t.example.com/a", "file.iso", 500, 512, pieces)

	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tf.PieceHashes) != 3 {
		t.Errorf("expected 3 piece hashes, got %d", len(tf.PieceHashes))
	}
	// Verify each piece hash was copied correctly
	for i, ph := range tf.PieceHashes {
		for j, b := range ph {
			if b != 'A' {
				t.Errorf("pieceHashes[%d][%d]: expected 'A', got %v", i, j, b)
			}
		}
	}
}

func TestNewTorrentFile_InfoHash(t *testing.T) {
	announce := "http://tracker.example.com/announce"
	name := "test.iso"
	length := 1000
	pieceLength := 256
	pieces := pieces20(1)

	input := buildTorrentBencode(announce, name, length, pieceLength, pieces)

	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Compute expected info hash independently
	info := map[string]any{
		"length":       length,
		"name":         name,
		"piece length": pieceLength,
		"pieces":       pieces,
	}
	rawInfo := parser.Encode(info)
	expectedHash := sha1.Sum([]byte(rawInfo))

	if tf.InfoHash != expectedHash {
		t.Errorf("InfoHash mismatch: expected %x, got %x", expectedHash, tf.InfoHash)
	}
}

func TestNewTorrentFile_MissingInfoDict(t *testing.T) {
	// No info key
	input := "d8:announce35:http://tracker.example.com/announcee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing info dict, got nil")
	}
}

func TestNewTorrentFile_MissingAnnounce(t *testing.T) {
	// info present but no announce
	input := "d4:infod6:lengthi100e4:name4:test12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing announce, got nil")
	}
}

func TestNewTorrentFile_MissingName(t *testing.T) {
	// info dict missing name
	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"length":       100,
			"piece length": 256,
			"pieces":       pieces20(1),
		},
	}
	input := parser.Encode(top)
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestNewTorrentFile_MissingLength(t *testing.T) {
	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.iso",
			"piece length": 256,
			"pieces":       pieces20(1),
		},
	}
	input := parser.Encode(top)
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing length, got nil")
	}
}

func TestNewTorrentFile_MissingPieceLength(t *testing.T) {
	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":   "test.iso",
			"length": 1000,
			"pieces": pieces20(1),
		},
	}
	input := parser.Encode(top)
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing piece length, got nil")
	}
}

func TestNewTorrentFile_MissingPieces(t *testing.T) {
	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.iso",
			"length":       1000,
			"piece length": 256,
		},
	}
	input := parser.Encode(top)
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing pieces, got nil")
	}
}

func TestNewTorrentFile_PiecesLengthNotMultipleOf20(t *testing.T) {
	// pieces string length not divisible by 20
	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.iso",
			"length":       1000,
			"piece length": 256,
			"pieces":       "AAAAAAAAAAAAAAAAAAA", // 19 bytes
		},
	}
	input := parser.Encode(top)
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for pieces length not multiple of 20, got nil")
	}
}

func TestNewTorrentFile_ZeroPieces(t *testing.T) {
	// Empty pieces string - length 0 is divisible by 20, so should succeed with 0 hashes
	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.iso",
			"length":       0,
			"piece length": 256,
			"pieces":       "",
		},
	}
	input := parser.Encode(top)
	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("unexpected error for empty pieces: %v", err)
	}
	if len(tf.PieceHashes) != 0 {
		t.Errorf("expected 0 piece hashes, got %d", len(tf.PieceHashes))
	}
}

func TestNewTorrentFile_PieceHashContents(t *testing.T) {
	// Verify that piece hash bytes are copied correctly, not just the count
	piece1 := strings.Repeat("A", 20)
	piece2 := strings.Repeat("B", 20)
	pieces := piece1 + piece2

	top := map[string]any{
		"announce": "http://tracker.example.com/announce",
		"info": map[string]any{
			"name":         "test.iso",
			"length":       1000,
			"piece length": 256,
			"pieces":       pieces,
		},
	}
	input := parser.Encode(top)
	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tf.PieceHashes) != 2 {
		t.Fatalf("expected 2 piece hashes, got %d", len(tf.PieceHashes))
	}
	for j := 0; j < 20; j++ {
		if tf.PieceHashes[0][j] != 'A' {
			t.Errorf("pieceHashes[0][%d]: expected 'A', got %v", j, tf.PieceHashes[0][j])
		}
		if tf.PieceHashes[1][j] != 'B' {
			t.Errorf("pieceHashes[1][%d]: expected 'B', got %v", j, tf.PieceHashes[1][j])
		}
	}
}