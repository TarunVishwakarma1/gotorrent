package torrent

import (
	"crypto/sha1"
	"strings"
	"testing"

	"github.com/tarunvishwakarma1/gotorret/parser"
)

// validBencode returns a well-formed bencoded torrent string for testing.
// The info dict keys in sorted order: length, name, "piece length", pieces.
func validBencode() string {
	return "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
}

func TestNewTorrentFile_Valid(t *testing.T) {
	tf, err := NewTorrentFile(validBencode())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tf.Announce != "http://tracker.example.com/announce" {
		t.Errorf("Announce = %q, want %q", tf.Announce, "http://tracker.example.com/announce")
	}
	if tf.Name != "test.iso" {
		t.Errorf("Name = %q, want %q", tf.Name, "test.iso")
	}
	if tf.Length != 1000 {
		t.Errorf("Length = %d, want 1000", tf.Length)
	}
	if tf.PieceLength != 256 {
		t.Errorf("PieceLength = %d, want 256", tf.PieceLength)
	}
}

func TestNewTorrentFile_PieceHashes_SinglePiece(t *testing.T) {
	tf, err := NewTorrentFile(validBencode())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tf.PieceHashes) != 1 {
		t.Fatalf("expected 1 piece hash, got %d", len(tf.PieceHashes))
	}
	// All 'A' bytes
	for i, b := range tf.PieceHashes[0] {
		if b != 'A' {
			t.Errorf("PieceHashes[0][%d] = %d, want %d ('A')", i, b, 'A')
		}
	}
}

func TestNewTorrentFile_PieceHashes_MultiplePieces(t *testing.T) {
	// 40 bytes = 2 pieces: first all 'A', second all 'B'
	pieces := strings.Repeat("A", 20) + strings.Repeat("B", 20)
	// Rebuild bencode with 40-char pieces
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi2000e4:name8:test.iso12:piece lengthi256e6:pieces40:" + pieces + "ee"
	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tf.PieceHashes) != 2 {
		t.Fatalf("expected 2 piece hashes, got %d", len(tf.PieceHashes))
	}
	for i := 0; i < 20; i++ {
		if tf.PieceHashes[0][i] != 'A' {
			t.Errorf("PieceHashes[0][%d] = %d, want 'A'", i, tf.PieceHashes[0][i])
		}
		if tf.PieceHashes[1][i] != 'B' {
			t.Errorf("PieceHashes[1][%d] = %d, want 'B'", i, tf.PieceHashes[1][i])
		}
	}
}

func TestNewTorrentFile_InfoHash(t *testing.T) {
	tf, err := NewTorrentFile(validBencode())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// Compute expected info hash manually
	infoDict := map[string]any{
		"length":       1000,
		"name":         "test.iso",
		"piece length": 256,
		"pieces":       strings.Repeat("A", 20),
	}
	rawInfo := parser.Encode(infoDict)
	expected := sha1.Sum([]byte(rawInfo))

	if tf.InfoHash != expected {
		t.Errorf("InfoHash = %x, want %x", tf.InfoHash, expected)
	}
}

func TestNewTorrentFile_MissingInfo(t *testing.T) {
	// Dict without "info" key
	input := "d8:announce35:http://tracker.example.com/announcee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing info dict, got nil")
	}
}

func TestNewTorrentFile_MissingAnnounce(t *testing.T) {
	// Dict without "announce" key
	input := "d4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing announce, got nil")
	}
}

func TestNewTorrentFile_MissingName(t *testing.T) {
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestNewTorrentFile_MissingLength(t *testing.T) {
	input := "d8:announce35:http://tracker.example.com/announce4:infod4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing length, got nil")
	}
}

func TestNewTorrentFile_MissingPieceLength(t *testing.T) {
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing piece length, got nil")
	}
}

func TestNewTorrentFile_MissingPieces(t *testing.T) {
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256eee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for missing pieces, got nil")
	}
}

func TestNewTorrentFile_InvalidPiecesLength(t *testing.T) {
	// 21 bytes is not a multiple of 20
	pieces := strings.Repeat("A", 21)
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces21:" + pieces + "ee"
	_, err := NewTorrentFile(input)
	if err == nil {
		t.Error("expected error for invalid pieces length, got nil")
	}
}

func TestNewTorrentFile_ZeroPieces(t *testing.T) {
	// Zero-length pieces string is valid (0 % 20 == 0), resulting in zero piece hashes
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi0e4:name8:test.iso12:piece lengthi256e6:pieces0:ee"
	tf, err := NewTorrentFile(input)
	if err != nil {
		t.Fatalf("expected no error for zero pieces, got: %v", err)
	}
	if len(tf.PieceHashes) != 0 {
		t.Errorf("expected 0 piece hashes, got %d", len(tf.PieceHashes))
	}
}