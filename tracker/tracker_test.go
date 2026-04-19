package tracker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tarunvishwakarma1/gotorret/parser"
	"github.com/tarunvishwakarma1/gotorret/torrent"
)

// makeTorrentFile creates a minimal TorrentFile for use in tracker tests.
func makeTorrentFile(announce string) *torrent.TorrentFile {
	var infoHash [20]byte
	copy(infoHash[:], "AAAAAAAAAAAAAAAAAAAA")
	var ph [20]byte
	copy(ph[:], "AAAAAAAAAAAAAAAAAAAA")
	return &torrent.TorrentFile{
		Announce:    announce,
		Name:        "test.iso",
		Length:      1000,
		PieceLength: 256,
		PieceHashes: [][20]byte{ph},
		InfoHash:    infoHash,
	}
}

// bencodePeersResponse returns a bencoded tracker response containing a "peers" key.
func bencodePeersResponse(peers string) string {
	return parser.Encode(map[string]any{
		"peers": peers,
	})
}

func TestGetPeersSuccess(t *testing.T) {
	expectedPeers := "PEER_DATA_12345678" // arbitrary compact peer string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, bencodePeersResponse(expectedPeers))
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peers != expectedPeers {
		t.Errorf("expected peers %q, got %q", expectedPeers, peers)
	}
}

func TestGetPeersMissingPeersKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Response has a dict but no "peers" key
		fmt.Fprint(w, parser.Encode(map[string]any{"interval": 1800}))
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for missing peers key, got nil")
	}
}

func TestGetPeersNonDictResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a bencoded integer instead of a dict
		fmt.Fprint(w, "i42e")
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for non-dict tracker response, got nil")
	}
}

func TestGetPeersInvalidURL(t *testing.T) {
	// A URL with a control character causes url.Parse to fail
	tf := makeTorrentFile("http://invalid\x00host/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for invalid announce URL, got nil")
	}
}

func TestGetPeersServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close the connection immediately to simulate network failure
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "not a hijacker", http.StatusInternalServerError)
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error when server closes connection, got nil")
	}
}

func TestGetPeersQueryParams(t *testing.T) {
	var capturedQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		fmt.Fprint(w, bencodePeersResponse("TESTPEERS123456789A"))
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	tf.Length = 5000
	_, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := capturedQuery
	// Check required parameters are present in the query string
	for _, param := range []string{"info_hash", "peer_id", "port", "uploaded", "downloaded", "left", "compact"} {
		if !containsParam(q, param) {
			t.Errorf("query string missing parameter %q: %s", param, q)
		}
	}
}

func containsParam(query, param string) bool {
	return len(query) > 0 && (len(query) >= len(param) && containsSubstring(query, param+"="))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestGetPeersPortIsFixed(t *testing.T) {
	var capturedPort string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPort = r.URL.Query().Get("port")
		fmt.Fprint(w, bencodePeersResponse("TESTPEERS123456789A"))
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	_, _ = GetPeers(tf)

	if capturedPort != "6881" {
		t.Errorf("expected port 6881, got %q", capturedPort)
	}
}

func TestGetPeersLeftEqualsTorrentLength(t *testing.T) {
	var capturedLeft string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedLeft = r.URL.Query().Get("left")
		fmt.Fprint(w, bencodePeersResponse("TESTPEERS123456789A"))
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	tf.Length = 99999
	_, _ = GetPeers(tf)

	if capturedLeft != "99999" {
		t.Errorf("expected left=99999, got %q", capturedLeft)
	}
}

func TestGetPeersCompactIsOne(t *testing.T) {
	var capturedCompact string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCompact = r.URL.Query().Get("compact")
		fmt.Fprint(w, bencodePeersResponse("TESTPEERS123456789A"))
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL)
	_, _ = GetPeers(tf)

	if capturedCompact != "1" {
		t.Errorf("expected compact=1, got %q", capturedCompact)
	}
}