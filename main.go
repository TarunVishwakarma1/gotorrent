package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	I    string = "i"
	DICT string = "d"
	E    string = "e"
	L    string = "l"
	SEP  string = ":"
)

func main() {
	torrent := "d8:announce35:http://tracker.example.com/announce"
	parse(torrent)

}

func parse(t string) {

	// making chan for testing
	ch := make(chan string)

	go func() {
		cc := <-ch
		fmt.Println(cc)
	}()

	for i := 0; i < len(t); i++ {
		if string(t[i]) == I {
			parseInteger(t, &i)
		}
		if string(t[i]) == DICT {
			parseDict(t, &i)
		}
	}
}

func parseInteger(t string, i *int) int {
	var intgr strings.Builder
	*i++
	for string(t[*i]) != E {
		intgr.WriteString(string(t[*i]))
		*i++
	}
	fmt.Println(intgr.String())
	if i, e := strconv.Atoi(intgr.String()); e != nil {
		fmt.Println("Failed to parse integer", intgr)
		return 0
	} else {
		return i
	}

}

func parseDict(t string, i *int) {
	dict := make(map[string]any)
	var k int
	var v int
	var key strings.Builder
	var value strings.Builder
	*i++
	dictKey(t, i, k, &key)

	*i++
	dictValue(t, i, v, &value)

	dict[key.String()] = value.String()

	fmt.Println("dict value:", dict)
}

func dictValue(t string, i *int, v int, value *strings.Builder) {
	for string(t[*i]) != SEP {
		v = v*10 + int(t[*i]-'0')
		*i++
	}

	for range v {
		*i++
		value.WriteString(string(t[*i]))
	}
}

func dictKey(t string, i *int, k int, key *strings.Builder) {
	for string(t[*i]) != SEP {
		k = k*10 + int(t[*i]-'0')
		*i++
	}

	for range k {
		*i++
		key.WriteString(string(t[*i]))

	}
}
