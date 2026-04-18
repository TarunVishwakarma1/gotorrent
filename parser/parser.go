package parser

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	iByte    byte = 'i'
	dictByte byte = 'd'
	eByte    byte = 'e'
	lByte    byte = 'l'
	sepByte  byte = ':'
)

// Parse will parse the torrent file string to any value
func Parse(t string) any {
	i := 0
	return parseValue(t, &i)
}

func parseValue(t string, i *int) any {
	if *i >= len(t) {
		return nil
	}
	switch t[*i] {
	case iByte:
		return parseInteger(t, i)
	case dictByte:
		return parseDict(t, i)
	case lByte:
		return parseList(t, i)
	default: // digit → string
		return parseString(t, i)
	}
}

func parseInteger(t string, i *int) int {
	var intgr strings.Builder
	*i++
	for *i < len(t) && t[*i] != eByte {
		intgr.WriteByte(t[*i])
		*i++
	}
	if res, e := strconv.Atoi(intgr.String()); e != nil {
		fmt.Println("Failed to parse integer", intgr.String())
		return 0
	} else {
		return res
	}
}

func parseDict(t string, i *int) map[string]any {
	dict := make(map[string]any)
	*i++

	for *i < len(t) && t[*i] != eByte {
		key := parseString(t, i)
		*i++

		value := parseValue(t, i)
		*i++

		dict[key] = value
	}
	return dict
}

func parseString(t string, i *int) string {
	var length int
	for *i < len(t) && t[*i] != sepByte {
		length = length*10 + int(t[*i]-'0')
		*i++
	}
	*i++

	var result strings.Builder
	for j := 0; j < length && *i < len(t); j++ {
		result.WriteByte(t[*i])
		*i++
	}
	*i--
	return result.String()
}

func parseList(t string, i *int) []any {
	var list []any
	*i++

	for *i < len(t) && t[*i] != eByte {
		value := parseValue(t, i)
		*i++
		list = append(list, value)
	}
	return list
}
