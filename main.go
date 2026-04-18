package main

import (
	"fmt"
	
	"github.com/tarunvishwakarma1/gotorret/parser"
)

func main() {
	torrent := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAA7:privatei1eee"
	result := parser.Parse(torrent)
	fmt.Println("Final result:", result)
}
