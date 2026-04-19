package client

import (
	"fmt"
	"net"
	"time"

	"github.com/tarunvishwakarma1/gotorret/handshake"
	"github.com/tarunvishwakarma1/gotorret/peers"
)

type Client struct {
	Conn     net.Conn
	InfoHash [20]byte
	PeerID   [20]byte
	Peer     peers.Peer
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

	fmt.Println("Protocol:", theirHs.Pstr)
	fmt.Printf("Reserved: %08b\n", theirHs.Reserved)
	fmt.Printf("InfoHash: %x\n", theirHs.InfoHash)
	fmt.Printf("PeerID:   %x\n", theirHs.PeerID)

	return &Client{
		Conn:     conn,
		InfoHash: infoHash,
		PeerID:   peerID,
		Peer:     peer,
	}, nil
}
