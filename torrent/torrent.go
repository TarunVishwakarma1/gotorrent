package torrent

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"

	"github.com/tarunvishwakarma1/gotorrent/parser"
)

// TorrentFileItem represents a single file within a multi-file torrent.
type TorrentFileItem struct {
	// Path is the relative path components for this file.
	Path []string
	// Length is the size of this file in bytes.
	Length int64
}

// TorrentFile holds parsed metadata from a .torrent file.
type TorrentFile struct {
	Announce    string
	Name        string
	Length      int
	PieceLength int
	PieceHashes [][20]byte
	InfoHash    [20]byte
	Port        uint16
	PeerID      [20]byte
	// Files holds per-file metadata for multi-file torrents.
	Files []TorrentFileItem
	// IsMultiFile is true when the torrent contains multiple files.
	IsMultiFile bool
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

	// Detect single-file vs multi-file torrent.
	var totalLength int
	var files []TorrentFileItem
	isMultiFile := false

	if rawFiles, hasFiles := info["files"]; hasFiles {
		// Multi-file torrent: info dict has "files" array instead of "length".
		isMultiFile = true
		fileList, ok := rawFiles.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid files list in multi-file torrent")
		}
		for _, f := range fileList {
			fmap, ok := f.(map[string]any)
			if !ok {
				continue
			}
			fileLen, _ := fmap["length"].(int)
			rawPath, _ := fmap["path"].([]any)
			pathParts := make([]string, 0, len(rawPath))
			for _, p := range rawPath {
				if s, ok := p.(string); ok {
					pathParts = append(pathParts, s)
				}
			}
			files = append(files, TorrentFileItem{
				Path:   pathParts,
				Length: int64(fileLen),
			})
			totalLength += fileLen
		}
	} else {
		// Single-file torrent: info dict has "length".
		length, ok := info["length"].(int)
		if !ok {
			return nil, fmt.Errorf("missing length")
		}
		totalLength = length
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
		Length:      totalLength,
		PieceLength: pieceLength,
		PieceHashes: pieceHashes,
		InfoHash:    infoHash,
		Port:        6881,
		PeerID:      peerID,
		Files:       files,
		IsMultiFile: isMultiFile,
	}, nil
}
