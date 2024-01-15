package lookup

import (
	"fmt"
	"net"
)

type HostTable struct {
	content map[string]net.Addr
}

func newHostTable() HostTable {
	return HostTable{map[string]net.Addr{}}
}

func (table HostTable) add(name string, addr net.Addr) {
	table.content[name] = addr
}

func (table HostTable) List() []string {
	hostnames := make([]string, 0, len(table.content))
	for hostname := range table.content {
		hostnames = append(hostnames, hostname)
	}
	return hostnames
}

func (table HostTable) Resolve(name string) (net.Addr, error) {
	addr, exists := table.content[name]
	if !exists {
		return nil, fmt.Errorf("could not resolve host %q", name)
	}
	return addr, nil
}
