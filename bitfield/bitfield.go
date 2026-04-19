package bitfield

import "fmt"

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

// String prints the bitfield in a human readable way for debugging
func (bf Bitfield) String() string {
	var result string
	for i := range len(bf) * 8 {
		if bf.HasPiece(i) {
			result += "1"
		} else {
			result += "0"
		}
		// add a space every 8 bits for readability
		if (i+1)%8 == 0 {
			result += " "
		}
	}
	return result
}

// HowMany returns how many pieces the peer has
func (bf Bitfield) HowMany() int {
	count := 0
	for i := range len(bf) * 8 {
		if bf.HasPiece(i) {
			count++
		}
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
