package main

import (
	"fmt"
	"os"

	"github.com/tarunvishwakarma1/gotorret/peers"
	"github.com/tarunvishwakarma1/gotorret/torrent"
	"github.com/tarunvishwakarma1/gotorret/tracker"
)

func main() {
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
}
