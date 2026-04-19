package main

import (
	"fmt"
	"os"

	"github.com/tarunvishwakarma1/gotorret/torrent"
)

// main reads "debian.torrent", constructs a TorrentFile from its contents, and prints the torrent's InfoHash in hexadecimal.
// Errors from reading the file or constructing the torrent are ignored (the file-read error can be overwritten and the constructor error is discarded), so the program may continue even if those operations fail.
func main() {
	tf := "debian.torrent"
	tdata, err := os.ReadFile(tf)
	t, err := torrent.NewTorrentFile(string(tdata))
	if err != nil {
		_ = fmt.Errorf("Error in making torrent file")
	}
	fmt.Printf("Final result: %x\n", t.InfoHash)
}
