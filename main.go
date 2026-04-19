package main

import (
	"fmt"
	"os"

	"github.com/tarunvishwakarma1/gotorret/client"
	"github.com/tarunvishwakarma1/gotorret/peers"
	"github.com/tarunvishwakarma1/gotorret/torrent"
	"github.com/tarunvishwakarma1/gotorret/tracker"
)

func main() {
	// 1. read the torrent file
	tdata, err := os.ReadFile("debian.torrent")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 2. parse it into a TorrentFile struct
	t, err := torrent.NewTorrentFile(string(tdata))
	if err != nil {
		fmt.Println("Error parsing torrent:", err)
		return
	}
	fmt.Printf("InfoHash: %x\n", t.InfoHash)
	fmt.Println("Name:", t.Name)
	fmt.Println("Length:", t.Length)

	// 3. get peers from tracker
	rawPeers, err := tracker.GetPeers(t)
	if err != nil {
		fmt.Println("Error getting peers:", err)
		return
	}

	// 4. decode peers
	peerList, err := peers.Decode(rawPeers)
	if err != nil {
		fmt.Println("Error decoding peers:", err)
		return
	}
	fmt.Println("Number of peers:", len(peerList))
	for _, p := range peerList {
		fmt.Println(p)
	}

	// 5. try handshake with peers until one succeeds
	var c *client.Client
	for _, p := range peerList {
		c, err = client.New(p, t.InfoHash, t.PeerID)
		if err != nil {
			fmt.Println("Failed to connect to peer:", p, "reason:", err)
			continue // try next peer
		}
		fmt.Println("Handshake successful with:", p)
		break // found a working peer, stop looking
	}

	if c == nil {
		fmt.Println("Could not connect to any peer")
		return
	}
	defer c.Conn.Close()
}
