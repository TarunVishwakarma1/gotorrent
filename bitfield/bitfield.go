package bitfield

import (
	"fmt"
	"math/bits"
	"strings"
)

// Bitfield represents which pieces a peer has
// each bit represents one piece
// 1 = has the piece, 0 = does not have the piece
type Bitfield []byte

// HasPiece checks if the peer has a specific piece
func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8

	// check bounds
	if byteIndex >= len(bf) {
		return false
	}

	// shift the bit we care about to the rightmost position
	// then AND with 1 to isolate it
	return bf[byteIndex]>>(7-offset)&1 != 0
}

// SetPiece marks a piece as downloaded in the bitfield
func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	// check bounds
	if byteIndex >= len(bf) {
		return
	}

	// shift 1 to the correct position and OR it in
	bf[byteIndex] |= 1 << (7 - offset)
}

// String prints the bitfield in a human readable way for debugging.
func (bf Bitfield) String() string {
	var sb strings.Builder
	sb.Grow(len(bf)*9) // 8 bits + 1 space per byte
	for i := range len(bf) * 8 {
		if bf.HasPiece(i) {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
		if (i+1)%8 == 0 {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

// HowMany returns how many pieces the peer has.
// Uses math/bits.OnesCount8 — maps to a single POPCNT hardware instruction
// on x86/arm64, making this O(n/8) instead of O(n).
func (bf Bitfield) HowMany() int {
	count := 0
	for _, b := range bf {
		count += bits.OnesCount8(b)
	}
	return count
}

// Validate checks that the bitfield is the right size
// numPieces is the total number of pieces in the torrent
func (bf Bitfield) Validate(numPieces int) error {
	// bitfield must have exactly enough bytes to hold all pieces
	expectedBytes := (numPieces + 7) / 8 // round up to nearest byte
	if len(bf) != expectedBytes {
		return fmt.Errorf("bitfield wrong size: got %d bytes expected %d bytes", len(bf), expectedBytes)
	}
	return nil
}
