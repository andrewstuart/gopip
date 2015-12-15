package relay

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

//Pip boy ports
const (
	UDPPort = 28000
	TCPPort = 27000
)

//Server represents a PipBoy game on a server
type Server struct {
	IP          net.IP
	IsBusy      bool
	MachineType string
}

var bcAddr = &net.UDPAddr{IP: net.IPv4bcast, Port: UDPPort}

//AutoDiscover is the command for autodiscovery
const AutoDiscover string = `{"cmd": "autodiscover"}`

//Discover returns a list of servers and their status
func (c *Client) Discover() ([]Server, error) {
	_, err := c.u.WriteToUDP([]byte(AutoDiscover), bcAddr)
	if err != nil {
		return nil, err
	}

	errC := make(chan error)
	srvC := make(chan Server)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("error discovering local ip addresses: %v", err)
	}

	localIPs := make([]net.IP, len(addrs))

	for i, addr := range addrs {
		switch addr := addr.(type) {
		case *net.IPAddr:
			localIPs[i] = addr.IP
		case *net.IPNet:
			localIPs[i] = addr.IP
		}
	}

	go func() {
		bs := make([]byte, 1024)
	readLoop:
		for {
			n, from, err := c.u.ReadFromUDP(bs)

			for _, ip := range localIPs {
				//Ignore local address
				if from.IP.Equal(ip) {
					continue readLoop
				}
			}

			if err != nil {
				select {
				case errC <- err:
				case <-time.After(10 * time.Millisecond):
				}
				return
			}

			var srv Server
			srv.IP = from.IP
			err = json.Unmarshal(bs[:n], &srv)
			if err != nil {
				select {
				case errC <- err:
				case <-time.After(10 * time.Millisecond):
				}
				return
			}

			select {
			case srvC <- srv:
			case <-time.After(10 * time.Millisecond):
				return
			}
		}
	}()

	servers := make([]Server, 0, 1)

	for {
		select {
		case <-time.After(250 * time.Millisecond):
			return servers, nil
		case s := <-srvC:
			servers = append(servers, s)
		case err := <-errC:
			return servers, err
		}
	}

}
