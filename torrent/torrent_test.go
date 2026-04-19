package torrent

import (
	"crypto/sha1"
	"strings"
	"testing"

	"github.com/tarunvishwakarma1/gotorrent/parser"
)

// buildTorrentBencode constructs a valid bencode torrent string from parts.
// pieces must be a multiple of 20 bytes.
func buildTorrentBencode(announce, name string, length, pieceLength int, pieces string) string {
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

// computeExpectedInfoHash returns the sha1 of the encoded info dict.
func computeExpectedInfoHash(name string, length, pieceLength int, pieces string) [20]byte {
	info := map[string]any{
		"length":       length,
		"name":         name,
		"piece length": pieceLength,
		"pieces":       pieces,
	}
	return sha1.Sum([]byte(parser.Encode(info)))
}

const (
	testAnnounce    = "http://example.com/announce"
	testName        = "test.iso"
	testLength      = 1000
	testPieceLength = 256
	// 20 bytes of 'A' — exactly one piece hash
	testPieces20 = "AAAAAAAAAAAAAAAAAAAA"
	// 40 bytes — two piece hashes
	testPieces40 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)

func TestNewTorrentFileValid(t *testing.T) {
	tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, testPieces20)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf.Announce != testAnnounce {
		t.Errorf("Announce = %q, want %q", tf.Announce, testAnnounce)
	}
	if tf.Name != testName {
		t.Errorf("Name = %q, want %q", tf.Name, testName)
	}
	if tf.Length != testLength {
		t.Errorf("Length = %d, want %d", tf.Length, testLength)
	}
	if tf.PieceLength != testPieceLength {
		t.Errorf("PieceLength = %d, want %d", tf.PieceLength, testPieceLength)
	}
}

func TestNewTorrentFilePieceHashesCount(t *testing.T) {
	t.Run("one piece", func(t *testing.T) {
		tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, testPieces20)
		tf, err := NewTorrentFile(tstr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tf.PieceHashes) != 1 {
			t.Errorf("PieceHashes len = %d, want 1", len(tf.PieceHashes))
		}
	})

	t.Run("two pieces", func(t *testing.T) {
		pieces40 := strings.Repeat("A", 40)
		tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, pieces40)
		tf, err := NewTorrentFile(tstr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tf.PieceHashes) != 2 {
			t.Errorf("PieceHashes len = %d, want 2", len(tf.PieceHashes))
		}
	})

	t.Run("five pieces", func(t *testing.T) {
		pieces100 := strings.Repeat("B", 100)
		tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, pieces100)
		tf, err := NewTorrentFile(tstr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tf.PieceHashes) != 5 {
			t.Errorf("PieceHashes len = %d, want 5", len(tf.PieceHashes))
		}
	})
}

func TestNewTorrentFilePieceHashContent(t *testing.T) {
	// First 20 bytes should match the first piece in the pieces string.
	pieces := strings.Repeat("X", 20) + strings.Repeat("Y", 20)
	tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, pieces)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var wantFirst [20]byte
	copy(wantFirst[:], strings.Repeat("X", 20))
	var wantSecond [20]byte
	copy(wantSecond[:], strings.Repeat("Y", 20))

	if tf.PieceHashes[0] != wantFirst {
		t.Errorf("PieceHashes[0] = %v, want %v", tf.PieceHashes[0], wantFirst)
	}
	if tf.PieceHashes[1] != wantSecond {
		t.Errorf("PieceHashes[1] = %v, want %v", tf.PieceHashes[1], wantSecond)
	}
}

func TestNewTorrentFileInfoHash(t *testing.T) {
	tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, testPieces20)
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := computeExpectedInfoHash(testName, testLength, testPieceLength, testPieces20)
	if tf.InfoHash != want {
		t.Errorf("InfoHash = %x, want %x", tf.InfoHash, want)
	}
}

func TestNewTorrentFileMissingInfo(t *testing.T) {
	// Torrent with no info key
	tstr := parser.Encode(map[string]any{
		"announce": testAnnounce,
	})
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for missing info, got nil")
	}
}

func TestNewTorrentFileMissingAnnounce(t *testing.T) {
	tstr := parser.Encode(map[string]any{
		"info": map[string]any{
			"length":       testLength,
			"name":         testName,
			"piece length": testPieceLength,
			"pieces":       testPieces20,
		},
	})
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for missing announce, got nil")
	}
}

func TestNewTorrentFileMissingName(t *testing.T) {
	tstr := parser.Encode(map[string]any{
		"announce": testAnnounce,
		"info": map[string]any{
			"length":       testLength,
			"piece length": testPieceLength,
			"pieces":       testPieces20,
		},
	})
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestNewTorrentFileMissingLength(t *testing.T) {
	tstr := parser.Encode(map[string]any{
		"announce": testAnnounce,
		"info": map[string]any{
			"name":         testName,
			"piece length": testPieceLength,
			"pieces":       testPieces20,
		},
	})
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for missing length, got nil")
	}
}

func TestNewTorrentFileMissingPieceLength(t *testing.T) {
	tstr := parser.Encode(map[string]any{
		"announce": testAnnounce,
		"info": map[string]any{
			"length": testLength,
			"name":   testName,
			"pieces": testPieces20,
		},
	})
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for missing piece length, got nil")
	}
}

func TestNewTorrentFileMissingPieces(t *testing.T) {
	tstr := parser.Encode(map[string]any{
		"announce": testAnnounce,
		"info": map[string]any{
			"length":       testLength,
			"name":         testName,
			"piece length": testPieceLength,
		},
	})
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for missing pieces, got nil")
	}
}

func TestNewTorrentFilePiecesNotMultipleOf20(t *testing.T) {
	// 19-byte pieces string is not a multiple of 20
	badPieces := strings.Repeat("A", 19)
	tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, badPieces)
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for pieces not multiple of 20, got nil")
	}
}

func TestNewTorrentFilePiecesLengthOneByte(t *testing.T) {
	// 1-byte pieces string — not a multiple of 20 — should error
	tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, "A")
	_, err := NewTorrentFile(tstr)
	if err == nil {
		t.Error("expected error for 1-byte pieces, got nil")
	}
}

func TestNewTorrentFileZeroPieces(t *testing.T) {
	// Empty pieces string is technically 0 % 20 == 0, so it should succeed with 0 hashes.
	tstr := buildTorrentBencode(testAnnounce, testName, testLength, testPieceLength, "")
	tf, err := NewTorrentFile(tstr)
	if err != nil {
		t.Fatalf("unexpected error for empty pieces: %v", err)
	}
	if len(tf.PieceHashes) != 0 {
		t.Errorf("PieceHashes len = %d, want 0", len(tf.PieceHashes))
	}
}
