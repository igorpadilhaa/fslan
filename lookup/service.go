package lookup

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

const multicastAddress = "239.255.45.23:5173"

var allHosts HostTable

func Start(server ServerDesc) (HostTable, error) {
	if allHosts.content != nil {
		return allHosts, nil
	}

	newTable := newHostTable()
	if err := searchHosts(server, newTable); err != nil {
		return HostTable{}, err
	}

	allHosts = newTable
	return allHosts, nil
}

func searchHosts(server ServerDesc, hosts HostTable) error {
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		return fmt.Errorf("address resolval: %w", err)
	}

	conn, err := net.ListenMulticastUDP(addr.Network(), nil, addr)
	if err != nil {
		return fmt.Errorf("could not open connection: %w", err)
	}

	if err := sendHi(conn, server); err != nil {
		return fmt.Errorf("could not send hi signal: %w", err)
	}

	go handleSignals(conn, server, hosts)
	return nil
}

type ServerDesc struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func sendHi(conn *net.UDPConn, server ServerDesc) error {
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		return err
	}

	msgData, err := json.Marshal(server)
	if err != nil {
		return err
	}

	return sendPacket(conn, addr, pktHi, msgData)
}

func sendPacket(conn *net.UDPConn, addr net.Addr, pktType packetType, data []byte) error {
	buffer := make([]byte, 0, 200)

	buffer = append(buffer, byte(pktHi))
	buffer = append(buffer, data...)

	_, err := conn.WriteTo(buffer, addr)
	return err
}

func handleSignals(conn *net.UDPConn, server ServerDesc, table HostTable) {
	buffer := make([]byte, 200)
	for {
		length, addr, err := conn.ReadFromUDP(buffer)
		if length == 0 && err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] discovery: %s", err)
		}

		receivedPkt, err := pack(buffer[:length])
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] discovery: %s", err)
		}

		switch receivedPkt.pktType {
		case pktHi:
			desc, err := parseDesc(receivedPkt)
			if err != nil {
				fmt.Fprintln(os.Stderr, "[ERROR] discovery: malformed message content")
				break
			}

			serverAddr := &net.TCPAddr{
				IP:   addr.IP,
				Port: desc.Port,
				Zone: "",
			}
			table.add(desc.Name, serverAddr)

			data, err := json.Marshal(server)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] discovery: failed to generate response: %s", err)
				break
			}

			if err := sendPacket(conn, addr, pktInfo, data); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] discovery: could not send response info: %s", err)
			}

		case pktInfo:
			desc, err := parseDesc(receivedPkt)
			if err != nil {
				fmt.Fprintln(os.Stderr, "[ERROR] discovery: malformed message content")
				break
			}

			serverAddr := &net.TCPAddr{
				IP:   addr.IP,
				Port: desc.Port,
				Zone: "",
			}
			table.add(desc.Name, serverAddr)

		default:
			fmt.Fprintf(os.Stderr, "unknwon packet type '%d' from %s", byte(receivedPkt.pktType), addr)
		}
	}
}

func parseDesc(receivedPkt packet) (ServerDesc, error) {
	var desc ServerDesc
	if err := json.Unmarshal(receivedPkt.data, &desc); err != nil {
		return ServerDesc{}, err
	}
	return desc, nil
}
