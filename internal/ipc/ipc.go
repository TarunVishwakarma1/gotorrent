// Package ipc implements single-instance enforcement via a local TCP socket.
//
// The first instance binds 127.0.0.1:19876. Subsequent instances detect the
// bound port, forward their torrent file path, and exit.
package ipc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const listenAddr = "127.0.0.1:19876"

// Server is the IPC listener held by the first application instance.
type Server struct {
	ln     net.Listener
	onFile func(path string)
}

// TryBecomeServer attempts to bind the IPC port.
// Returns a Server if this is the first instance, nil if another instance
// is already running. Any other error is returned as-is.
func TryBecomeServer(onFile func(path string)) (*Server, error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		// Port already bound — another instance is running.
		return nil, nil //nolint:nilerr
	}
	s := &Server{ln: ln, onFile: onFile}
	go s.serve()
	return s, nil
}

// IsRunning returns true if another instance is already listening.
func IsRunning() bool {
	conn, err := net.Dial("tcp", listenAddr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// SendPath forwards a torrent file path to the running instance.
func SendPath(path string) error {
	conn, err := net.Dial("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("ipc: connect: %w", err)
	}
	defer conn.Close()
	_, err = fmt.Fprintf(conn, "%s\n", path)
	return err
}

// Close shuts down the IPC listener.
func (s *Server) Close() {
	s.ln.Close()
}

func (s *Server) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" && s.onFile != nil {
			s.onFile(path)
		}
	}
}
