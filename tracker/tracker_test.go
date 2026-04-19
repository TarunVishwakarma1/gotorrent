package tracker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tarunvishwakarma1/gotorret/torrent"
)

// newTestTorrentFile creates a TorrentFile with the given announce URL for testing.
func newTestTorrentFile(announceURL string) *torrent.TorrentFile {
	return &torrent.TorrentFile{
		Announce:    announceURL,
		Name:        "test.iso",
		Length:      1000,
		PieceLength: 256,
		PieceHashes: [][20]byte{{}},
		InfoHash:    [20]byte{},
	}
}

func TestGetPeers_Success(t *testing.T) {
	// "ABCDEF" is a 6-byte compact peer representation
	peersData := "ABCDEF"
	responseBody := fmt.Sprintf("d5:peers%d:%se", len(peersData), peersData)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if peers != peersData {
		t.Errorf("GetPeers returned %q, want %q", peers, peersData)
	}
}

func TestGetPeers_QueryParamsPresent(t *testing.T) {
	peersData := "ABCDEF"
	responseBody := fmt.Sprintf("d5:peers%d:%se", len(peersData), peersData)
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	_, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	requiredParams := []string{"info_hash", "peer_id", "port", "uploaded", "downloaded", "left", "compact"}
	for _, param := range requiredParams {
		if !strings.Contains(capturedQuery, param) {
			t.Errorf("expected query param %q in request, not found in %q", param, capturedQuery)
		}
	}
}

func TestGetPeers_PortIs6881(t *testing.T) {
	peersData := "ABCDEF"
	responseBody := fmt.Sprintf("d5:peers%d:%se", len(peersData), peersData)
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	_, _ = GetPeers(tf)

	if !strings.Contains(capturedQuery, "port=6881") {
		t.Errorf("expected port=6881 in query, got: %q", capturedQuery)
	}
}

func TestGetPeers_LengthInQueryParams(t *testing.T) {
	peersData := "ABCDEF"
	responseBody := fmt.Sprintf("d5:peers%d:%se", len(peersData), peersData)
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	tf.Length = 54321
	_, _ = GetPeers(tf)

	if !strings.Contains(capturedQuery, "left=54321") {
		t.Errorf("expected left=54321 in query, got: %q", capturedQuery)
	}
}

func TestGetPeers_InvalidAnnounceURL(t *testing.T) {
	tf := newTestTorrentFile("://this-is-not-a-valid-url")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for invalid announce URL, got nil")
	}
}

func TestGetPeers_InvalidTrackerResponse_NotBencoded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "this is not bencoded data at all")
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for non-bencoded tracker response, got nil")
	}
}

func TestGetPeers_MissingPeersField(t *testing.T) {
	// Valid bencoded dict but without "peers" key
	responseBody := "d8:intervali1800ee"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error for missing peers field, got nil")
	}
}

func TestGetPeers_PeersFieldNotString(t *testing.T) {
	// "peers" is present but as a list (non-compact format), not a string
	responseBody := "d5:peersli1eee"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error when peers field is not a string, got nil")
	}
}

func TestGetPeers_ServerUnavailable(t *testing.T) {
	// Point to a port that is not listening
	tf := newTestTorrentFile("http://127.0.0.1:1/announce")
	_, err := GetPeers(tf)
	if err == nil {
		t.Error("expected error when server is unavailable, got nil")
	}
}

func TestGetPeers_EmptyPeers(t *testing.T) {
	// Empty peers string is valid (zero peers)
	responseBody := "d5:peers0:e"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	tf := newTestTorrentFile(server.URL + "/announce")
	peers, err := GetPeers(tf)
	if err != nil {
		t.Fatalf("expected no error for empty peers, got: %v", err)
	}
	if peers != "" {
		t.Errorf("expected empty peers string, got %q", peers)
	}
}