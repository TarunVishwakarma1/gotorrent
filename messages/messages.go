package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

// message IDs
const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8
)

// Message represents a BitTorrent message
type Message struct {
	ID      uint8
	Payload []byte
}

// Serialize turns a message into bytes to send over TCP
// format: [4 bytes length][1 byte ID][variable bytes payload]
// nil message means keepalive → [0 0 0 0]
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4) // keepalive
	}

	length := 1 + len(m.Payload) // 1 for ID + payload length
	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], uint32(length)) // 4 bytes length
	buf[4] = m.ID                                        // 1 byte ID
	copy(buf[5:], m.Payload)                             // rest is payload

	return buf
}

// Read parses a message from a stream
// format: [4 bytes length][1 byte ID][variable bytes payload]
func Read(r io.Reader) (*Message, error) {
	// step 1: read the 4 byte length prefix
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// step 2: keepalive message has length 0
	if length == 0 {
		return nil, nil // nil means keepalive
	}

	// step 3: read the rest of the message
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read message body: %w", err)
	}

	// step 4: first byte is ID, rest is payload
	return &Message{
		ID:      messageBuf[0],
		Payload: messageBuf[1:],
	}, nil
}

// --- building messages to send ---

// NewInterested creates an Interested message
// no payload needed
func NewInterested() *Message {
	return &Message{
		ID: MsgInterested,
	}
}

// NewNotInterested creates a NotInterested message
// no payload needed
func NewNotInterested() *Message {
	return &Message{
		ID: MsgNotInterested,
	}
}

// NewUnchoke creates an Unchoke message
// no payload needed
func NewUnchoke() *Message {
	return &Message{
		ID: MsgUnchoke,
	}
}

// NewHave creates a Have message
// payload: [4 bytes piece index]
func NewHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{
		ID:      MsgHave,
		Payload: payload,
	}
}

// NewRequest creates a Request message
// payload: [4 bytes piece index][4 bytes offset][4 bytes length]
func NewRequest(index, offset, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))   // which piece
	binary.BigEndian.PutUint32(payload[4:8], uint32(offset))  // byte offset within piece
	binary.BigEndian.PutUint32(payload[8:12], uint32(length)) // how many bytes
	return &Message{
		ID:      MsgRequest,
		Payload: payload,
	}
}

// NewCancel creates a Cancel message
// same payload format as Request
func NewCancel(index, offset, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{
		ID:      MsgCancel,
		Payload: payload,
	}
}

// --- parsing messages we receive ---

// ParseHave parses the payload of a Have message
// returns the piece index
func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected Have (ID 4) got ID %d", msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4 got %d", len(msg.Payload))
	}
	index := binary.BigEndian.Uint32(msg.Payload)
	return int(index), nil
}

// ParsePiece parses the payload of a Piece message
// copies the block data into buf at the right offset
// returns number of bytes copied
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected Piece (ID 7) got ID %d", msg.ID)
	}
	// payload: [4 bytes index][4 bytes offset][variable data]
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short: %d", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected piece index %d got %d", index, parsedIndex)
	}

	offset := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	data := msg.Payload[8:]

	if offset+len(data) > len(buf) {
		return 0, fmt.Errorf("data out of bounds: offset %d + length %d > buf %d", offset, len(data), len(buf))
	}

	copy(buf[offset:], data)
	return len(data), nil
}

// String returns a human readable message name for debugging
func (m *Message) String() string {
	if m == nil {
		return "keepalive"
	}
	names := map[uint8]string{
		MsgChoke:         "Choke",
		MsgUnchoke:       "Unchoke",
		MsgInterested:    "Interested",
		MsgNotInterested: "NotInterested",
		MsgHave:          "Have",
		MsgBitfield:      "Bitfield",
		MsgRequest:       "Request",
		MsgPiece:         "Piece",
		MsgCancel:        "Cancel",
	}
	name, ok := names[m.ID]
	if !ok {
		name = "Unknown"
	}
	return fmt.Sprintf("Message{ID: %s, Payload: %d bytes}", name, len(m.Payload))
}
