package lookup

import "errors"

type packetType byte

const (
	pktHi packetType = 1 << iota
	pktInfo
)

type packet struct {
	pktType packetType
	data    []byte
}

func pack(data []byte) (packet, error) {
	if len(data) == 0 {
		return packet{}, errors.New("malformed packet")
	}

	pktType := packetType(data[0])
	return packet{pktType, data[1:]}, nil
}
