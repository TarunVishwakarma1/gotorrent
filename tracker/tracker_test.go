package tracker

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tarunvishwakarma1/gotorrent/parser"
	"github.com/tarunvishwakarma1/gotorrent/torrent"
)

// makeTorrentFile returns a minimal TorrentFile with the given announce URL.
func makeTorrentFile(announce string) *torrent.TorrentFile {
	return &torrent.TorrentFile{
		Announce:    announce,
		Name:        "test.iso",
		Length:      1000,
		PieceLength: 256,
		PieceHashes: [][20]byte{},
	}
}

// bencodeResponse encodes a map as a bencode string.
func bencodeResponse(m map[string]any) string {
	return parser.Encode(m)
}

// startServer creates and starts an httptest.Server, skipping the test if the
// loopback TCP stack is not reachable in the current environment.
func startServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		srv.Close()
		t.Skip("loopback TCP unavailable in this environment; skipping network test")
	}
	conn.Close()
	return srv
}

func TestGetPeersSuccess(t *testing.T) {
	// Compact peers: 6 bytes per peer (4 IP + 2 port), two peers = 12 bytes
	rawPeers := "\x7f\x00\x00\x01\x1a\xe1\xc0\xa8\x01\x01\x1a\xe1"
	body := bencodeResponse(map[string]any{
		"peers":    rawPeers,
		"interval": 1800,
	})

	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peers != rawPeers {
		t.Errorf("peers = %q, want %q", peers, rawPeers)
	}
}

func TestGetPeersInvalidAnnounceURL(t *testing.T) {
	// url.Parse does not error on most strings, but a completely invalid scheme triggers an error
	tf := makeTorrentFile("://invalid-url")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for invalid announce URL, got nil")
	}
}

func TestGetPeersHTTPRequestFails(t *testing.T) {
	// Create a server, close it immediately so the HTTP request fails.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL + "/announce"
	srv.Close()

	tf := makeTorrentFile(url)
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for failed HTTP request, got nil")
	}
}

func TestGetPeersNonDictResponse(t *testing.T) {
	// Tracker returns a bencode integer — not a dict.
	body := parser.Encode(42)

	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for non-dict tracker response, got nil")
	}
}

func TestGetPeersMissingPeersKey(t *testing.T) {
	// Valid dict but no "peers" key.
	body := bencodeResponse(map[string]any{
		"interval": 1800,
	})

	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for missing peers key, got nil")
	}
}

func TestGetPeersTrackerParams(t *testing.T) {
	// Verify that required query parameters are present in the tracker request.
	rawPeers := "\x7f\x00\x00\x01\x1a\xe1"
	body := bencodeResponse(map[string]any{"peers": rawPeers})

	var capturedQuery string
	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	tf.Length = 5000
	_, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []string{"info_hash", "peer_id", "port", "uploaded", "downloaded", "left", "compact"}
	for _, param := range expectedParams {
		if !containsParam(capturedQuery, param) {
			t.Errorf("expected query param %q in %q", param, capturedQuery)
		}
	}
}

func TestGetPeersPortIsFixed(t *testing.T) {
	// The port parameter must always be 6881.
	rawPeers := "\x7f\x00\x00\x01\x1a\xe1"
	body := bencodeResponse(map[string]any{"peers": rawPeers})

	var capturedQuery string
	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsParam(capturedQuery, "port") {
		t.Errorf("port param missing from query %q", capturedQuery)
	}
	if !containsParamValue(capturedQuery, "port", "6881") {
		t.Errorf("port param in %q is not 6881", capturedQuery)
	}
}

func TestGetPeersCompactIsOne(t *testing.T) {
	rawPeers := "\x7f\x00\x00\x01\x1a\xe1"
	body := bencodeResponse(map[string]any{"peers": rawPeers})

	var capturedQuery string
	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsParamValue(capturedQuery, "compact", "1") {
		t.Errorf("compact param in %q is not 1", capturedQuery)
	}
}

func TestGetPeersEmptyPeers(t *testing.T) {
	// Tracker returns empty peers string — should succeed and return "".
	body := bencodeResponse(map[string]any{"peers": ""})

	srv := startServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peers != "" {
		t.Errorf("peers = %q, want empty string", peers)
	}
}

// containsParam checks if the raw query string contains a given parameter name.
func containsParam(rawQuery, param string) bool {
	for _, kv := range splitQuery(rawQuery) {
		if kv == param || len(kv) > len(param) && kv[:len(param)+1] == param+"=" {
			return true
		}
	}
	return false
}

// containsParamValue checks if rawQuery contains param=value.
func containsParamValue(rawQuery, param, value string) bool {
	target := param + "=" + value
	for _, kv := range splitQuery(rawQuery) {
		if kv == target {
			return true
		}
	}
	return false
}

func splitQuery(rawQuery string) []string {
	if rawQuery == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i <= len(rawQuery); i++ {
		if i == len(rawQuery) || rawQuery[i] == '&' {
			parts = append(parts, rawQuery[start:i])
			start = i + 1
		}
	}
	return parts
}
