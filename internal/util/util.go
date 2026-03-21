// Package util contains some helper functions
package util

import (
	"fmt"
	"net"
)

func GetIfaceByIP(IPAddr net.IP) (*net.Interface, error) {
	allIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range allIfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			addr, ok := addr.(*net.IPNet)
			if !ok {
				return nil, fmt.Errorf("error parsing Interface IP address")
			}
			if addr.Contains(IPAddr) {
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("no interface connected to that network")
}
