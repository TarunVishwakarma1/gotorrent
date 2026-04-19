package torrent

import (
	"crypto/rand"
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
	Port        uint16
	PeerID      [20]byte
}

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

	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate peer id: %w", err)
	}

	return &TorrentFile{
		Announce:    announce,
		Name:        name,
		Length:      length,
		PieceLength: pieceLength,
		PieceHashes: pieceHashes,
		InfoHash:    infoHash,
		Port:        6881,
		PeerID:      peerID,
	}, nil
}
