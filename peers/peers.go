package peers

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

func Decode(rawPeers string) ([]Peer, error) {
	var peers []Peer
	err := checkBlob(rawPeers)
	if err != nil {
		return nil, err
	}

	for i := range len(rawPeers) / 6 {
		start := i * 6
		ip := net.IP(rawPeers[start : start+4])
		port := binary.BigEndian.Uint16([]byte(rawPeers[start+4 : start+6]))
		peer := Peer{
			IP:   ip,
			Port: port,
		}
		peers = append(peers, peer)
	}
	return peers, nil
}

func checkBlob(rawPeers string) error {
	if len(rawPeers)%6 != 0 {
		return fmt.Errorf("Error getting peers. URL is malformed. URL length: %d", len(rawPeers))
	}
	return nil
}
