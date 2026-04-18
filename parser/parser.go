package parser

import "fmt"

// Parse will parse the torrent file string to any value
func Parse(t string) any {
	fmt.Println("Torrent String", t)
	return t
}
