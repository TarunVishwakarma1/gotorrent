package tracker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tarunvishwakarma1/gotorret/parser"
	"github.com/tarunvishwakarma1/gotorret/torrent"
)

// makeTorrentFile builds a minimal TorrentFile with the given announce URL.
func makeTorrentFile(announce string) *torrent.TorrentFile {
	return &torrent.TorrentFile{
		Announce:    announce,
		Name:        "test.iso",
		Length:      1000,
		PieceLength: 256,
		PieceHashes: [][20]byte{},
		InfoHash:    [20]byte{},
	}
}

// bencodeDict returns a bencoded dict string with the given key-value pairs.
func bencodeResponse(m map[string]any) string {
	return parser.Encode(m)
}

func TestGetPeers_Success(t *testing.T) {
	rawPeers := strings.Repeat("X", 12) // 2 peers × 6 bytes each (compact format)
	respBody := bencodeResponse(map[string]any{
		"interval": 1800,
		"peers":    rawPeers,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peers != rawPeers {
		t.Errorf("expected peers %q, got %q", rawPeers, peers)
	}
}

func TestGetPeers_TrackerReturnsEmptyPeers(t *testing.T) {
	respBody := bencodeResponse(map[string]any{
		"interval": 1800,
		"peers":    "",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("unexpected error for empty peers: %v", err)
	}
	if peers != "" {
		t.Errorf("expected empty peers, got %q", peers)
	}
}

func TestGetPeers_MissingPeersKey(t *testing.T) {
	// Tracker responds with a dict that has no "peers" key
	respBody := bencodeResponse(map[string]any{
		"interval": 1800,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for missing peers key, got nil")
	}
}

func TestGetPeers_InvalidTrackerResponse_NotBencode(t *testing.T) {
	// Tracker responds with non-bencoded body
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not bencode at all")
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for non-bencode tracker response, got nil")
	}
}

func TestGetPeers_InvalidAnnounceURL(t *testing.T) {
	// "://bad" is an invalid URL that url.Parse will reject
	tf := makeTorrentFile("://bad")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for invalid announce URL, got nil")
	}
}

func TestGetPeers_ServerUnreachable(t *testing.T) {
	// Point to a server that is closed
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := srv.URL
	srv.Close() // close before request

	tf := makeTorrentFile(addr + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for unreachable server, got nil")
	}
}

func TestGetPeers_RequestContainsRequiredParams(t *testing.T) {
	// Verify the tracker receives expected query parameters
	var gotQuery string
	respBody := bencodeResponse(map[string]any{"peers": ""})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	tf.Length = 5000
	_, _ = GetPeers(tf)

	requiredParams := []string{"info_hash", "peer_id", "port", "uploaded", "downloaded", "left", "compact"}
	for _, param := range requiredParams {
		if !strings.Contains(gotQuery, param) {
			t.Errorf("expected query param %q in request, got: %s", param, gotQuery)
		}
	}
}

func TestGetPeers_RequestUsesCompact1(t *testing.T) {
	respBody := bencodeResponse(map[string]any{"peers": ""})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		compact := r.URL.Query().Get("compact")
		if compact != "1" {
			t.Errorf("expected compact=1, got %q", compact)
		}
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, _ = GetPeers(tf)
}

func TestGetPeers_RequestUsesPort6881(t *testing.T) {
	respBody := bencodeResponse(map[string]any{"peers": ""})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		port := r.URL.Query().Get("port")
		if port != "6881" {
			t.Errorf("expected port=6881, got %q", port)
		}
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	_, _ = GetPeers(tf)
}

func TestGetPeers_LeftParamMatchesFileLength(t *testing.T) {
	respBody := bencodeResponse(map[string]any{"peers": ""})
	expectedLeft := "98765"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		left := r.URL.Query().Get("left")
		if left != expectedLeft {
			t.Errorf("expected left=%s, got %q", expectedLeft, left)
		}
		fmt.Fprint(w, respBody)
	}))
	defer srv.Close()

	tf := makeTorrentFile(srv.URL + "/announce")
	tf.Length = 98765
	_, _ = GetPeers(tf)
}