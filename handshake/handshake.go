package handshake

import (
	"fmt"
	"io"
)

const pstr = "BitTorrent protocol"

type Handshake struct {
	Pstr     string
	Reserved [8]byte // extension bits
	InfoHash [20]byte
	PeerID   [20]byte
}

func New(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     pstr,
		Reserved: [8]byte{}, // all zeros for now
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 68)
	buf[0] = byte(len(h.Pstr))
	copy(buf[1:20], h.Pstr)
	copy(buf[20:28], h.Reserved[:]) // use reserved bytes instead of hardcoded zeros
	copy(buf[28:48], h.InfoHash[:])
	copy(buf[48:68], h.PeerID[:])
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read pstr length: %w", err)
	}
	pstrLen := int(lengthBuf[0])
	if pstrLen == 0 {
		return nil, fmt.Errorf("pstr length is zero")
	}

	remainingBuf := make([]byte, pstrLen+48)
	_, err = io.ReadFull(r, remainingBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}

	if string(remainingBuf[0:pstrLen]) != pstr {
		return nil, fmt.Errorf("wrong protocol: %s", string(remainingBuf[0:pstrLen]))
	}

	var reserved [8]byte
	copy(reserved[:], remainingBuf[pstrLen:pstrLen+8])

	var infoHash, peerID [20]byte
	copy(infoHash[:], remainingBuf[pstrLen+8:pstrLen+28])
	copy(peerID[:], remainingBuf[pstrLen+28:pstrLen+48])

	return &Handshake{
		Pstr:     pstr,
		Reserved: reserved,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}

func (h *Handshake) Verify(infoHash [20]byte) error {
	if h.InfoHash != infoHash {
		return fmt.Errorf("infohash mismatch: expected %x got %x", infoHash, h.InfoHash)
	}
	return nil
}
