package client

import (
	"fmt"
	"net"
	"time"

	"github.com/tarunvishwakarma1/gotorrent/bitfield"
	"github.com/tarunvishwakarma1/gotorrent/handshake"
	message "github.com/tarunvishwakarma1/gotorrent/messages"
	"github.com/tarunvishwakarma1/gotorrent/peers"
)

type Client struct {
	Conn     net.Conn
	InfoHash [20]byte
	PeerID   [20]byte
	Peer     peers.Peer
	Choked   bool
	Bitfield bitfield.Bitfield
}

func New(peer peers.Peer, infoHash [20]byte, peerID [20]byte) (*Client, error) {
	// step 1: open TCP connection with 3 second timeout
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %w", peer, err)
	}

	// step 2: send our handshake
	hs := handshake.New(infoHash, peerID)
	_, err = conn.Write(hs.Serialize())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send handshake: %w", err)
	}

	// step 3: read their handshake back
	theirHs, err := handshake.Read(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}

	// step 4: verify infohash matches
	err = theirHs.Verify(infoHash)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("infohash mismatch: %w", err)
	}

	// Read messages until we get a bitfield; skip keepalives (nil messages).
	var msg *message.Message
	for {
		msg, err = message.Read(conn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to read bitfield: %w", err)
		}
		if msg == nil {
			// keepalive — try again
			continue
		}
		break
	}
	if msg.ID != message.MsgBitfield {
		conn.Close()
		return nil, fmt.Errorf("expected bitfield got ID %d", msg.ID)
	}

	fmt.Println("Protocol:", theirHs.Pstr)
	fmt.Printf("Reserved: %08b\n", theirHs.Reserved)
	fmt.Printf("InfoHash: %x\n", theirHs.InfoHash)
	fmt.Printf("PeerID:   %x\n", theirHs.PeerID)

	return &Client{
		Conn:     conn,
		InfoHash: infoHash,
		PeerID:   peerID,
		Peer:     peer,
		Choked:   true,                           // ← choked by default
		Bitfield: bitfield.Bitfield(msg.Payload), // ← store their bitfield
	}, nil
}
