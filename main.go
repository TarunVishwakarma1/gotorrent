package main

import (
	"fmt"
	"os"

	"github.com/tarunvishwakarma1/gotorret/torrent"
)

func main() {
	tf := "debian.torrent"
	tdata, err := os.ReadFile(tf)
	t, err := torrent.NewTorrentFile(string(tdata))
	if err != nil {
		_ = fmt.Errorf("Error in making torrent file")
	}
	fmt.Printf("Final result: %x\n", t.InfoHash)
}
