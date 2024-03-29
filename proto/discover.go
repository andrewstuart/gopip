package proto

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Pip boy ports
const (
	UDPPort = 28000
	TCPPort = 27000
)

// Server represents a PipBoy game on a server
type Server struct {
	Address     string
	IsBusy      bool
	MachineType string
}

var bcAddr = &net.UDPAddr{IP: net.IPv4bcast, Port: UDPPort}

// AutoDiscover is the command for autodiscovery
const AutoDiscover string = `{"cmd": "autodiscover"}`

// Discover returns a list of servers and their status
func Discover(ctx context.Context) ([]Server, error) {
	servers := make([]Server, 0, 1)

	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: UDPPort + 1}

	l, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return servers, err
	}
	defer l.Close()

	_, err = l.WriteToUDP([]byte(AutoDiscover), bcAddr)
	if err != nil {
		return servers, err
	}

	errC := make(chan error)
	srvC := make(chan Server)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return servers, fmt.Errorf("error discovering local ip addresses: %v", err)
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
		defer close(errC)
		defer close(srvC)

		bs := make([]byte, 1024)
	readLoop:
		for {
			n, from, err := l.ReadFromUDP(bs)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errC <- err:
				case <-time.After(10 * time.Millisecond):
				}
				return
			}

			for _, ip := range localIPs {
				//Ignore local address
				if from.IP.Equal(ip) {
					continue readLoop
				}
			}

			if err != nil {
				select {
				case <-ctx.Done():
				case errC <- err:
				case <-time.After(10 * time.Millisecond):
				}
				return
			}

			var srv Server
			srv.Address = from.IP.String()
			err = json.Unmarshal(bs[:n], &srv)
			if err != nil {
				select {
				case <-ctx.Done():
				case errC <- err:
				case <-time.After(10 * time.Millisecond):
				}
				return
			}

			select {
			case <-ctx.Done():
			case srvC <- srv:
			case <-time.After(10 * time.Millisecond):
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
		case s := <-srvC:
			servers = append(servers, s)
			return servers, nil
		case err := <-errC:
			return servers, err
		}
	}
}
