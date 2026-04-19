package tracker

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tarunvishwakarma1/gotorret/parser"
	"github.com/tarunvishwakarma1/gotorret/torrent"
)

// GetPeers contacts the torrent's announce tracker and retrieves the raw compact peer list.
// It returns the compact peer list as provided by the tracker, or an error if the announce URL is invalid, the HTTP request or response read fails, the tracker response cannot be decoded, or the peers field is absent.
func GetPeers(tf *torrent.TorrentFile) (string, error) {
	var peerId [20]byte
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", fmt.Errorf("Failed to parse announce URL")
	}
	_, err = rand.Read(peerId[:])
	if err != nil {
		return "", fmt.Errorf("failed to generate peer id: %w", err)
	}

	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{"6881"},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(tf.Length)},
		"compact":    []string{"1"},
	}

	base.RawQuery = params.Encode()
	trackerUrl := base.String()

	resp, err := http.Get(trackerUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	decoded, ok := parser.Decode(string(body)).(map[string]any)
	if !ok {
		return "", fmt.Errorf("failed to decode tracker response")
	}
	rawPeers, ok := decoded["peers"].(string)
	if !ok {
		return "", fmt.Errorf("missing peers in tracker response")
	}
	return rawPeers, nil

}
