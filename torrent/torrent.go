package torrent

import (
	"crypto/sha1"
	"fmt"

	"github.com/tarunvishwakarma1/gotorret/parser"
)

type TorrentFile struct {
	Announce    string
	Name        string
	Length      int
	PieceLength int
	PieceHashes [][20]byte
	InfoHash    [20]byte
}

// NewTorrentFile parses a bencoded torrent string and constructs a TorrentFile containing its decoded metadata.
// 
// It extracts the top-level "announce" URL and the "info" dictionary fields: "name", "length", "piece length"
// and "pieces". The "pieces" value is split into consecutive 20-byte SHA-1 piece hashes and stored as a slice
// of [20]byte. The function also computes InfoHash as the SHA-1 of the bencoded info dictionary.
// 
// If any required field is missing or has an unexpected type, or if the pieces data length is not a multiple
// of 20, an error describing the problem is returned.
func NewTorrentFile(tstr string) (*TorrentFile, error) {
	m := parser.Decode(tstr).(map[string]any)

	info, ok := m["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing info dict")
	}

	announce, ok := m["announce"].(string)
	if !ok {
		return nil, fmt.Errorf("missing announce")
	}

	name, ok := info["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing name")
	}

	length, ok := info["length"].(int)
	if !ok {
		return nil, fmt.Errorf("missing length")
	}

	pieceLength, ok := info["piece length"].(int)
	if !ok {
		return nil, fmt.Errorf("missing piece length")
	}

	pieces, ok := info["pieces"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pieces")
	}
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("Hash Length not correct, File is corrupted")
	}

	numPieces := len(pieces) / 20
	pieceHashes := make([][20]byte, numPieces)

	for i := range numPieces {
		start := i * 20
		end := start + 20
		copy(pieceHashes[i][:], pieces[start:end])
	}

	rawInfo := parser.Encode(info)
	infoHash := sha1.Sum([]byte(rawInfo))

	return &TorrentFile{
		Announce:    announce,
		Name:        name,
		Length:      length,
		PieceLength: pieceLength,
		PieceHashes: pieceHashes,
		InfoHash:    infoHash,
	}, nil
}
