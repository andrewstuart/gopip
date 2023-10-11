package relay

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/andrewstuart/gopip/proto"
	"github.com/andrewstuart/multierrgroup"
)

var ErrorNoServer = fmt.Errorf("no server found")

type relay struct {
	srv *proto.Server
}

func (r relay) listenUDP(ctx context.Context) error {
	l, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: proto.UDPPort})
	if err != nil {
		return fmt.Errorf("error listening on udp: %w", err)
	}

	localAddr := l.LocalAddr().(*net.UDPAddr)

	bs := make([]byte, 1024)
	for {
		n, addr, err := l.ReadFrom(bs)
		if err != nil {
			return fmt.Errorf("error reading from udp: %w", err)
		}

		m := make(map[string]any)
		json.Unmarshal(bs[:n], &m)
		fmt.Printf("map %#v\n", m)

		add, ok := addr.(*net.UDPAddr)
		if !ok {
			return fmt.Errorf("error converting addr to udpaddr")
		}
		// if this is the known server, ignore it
		if add.IP.String() == r.srv.Address || add.IP.String() == localAddr.IP.String() {
			continue
		}

		// pretend to be the server
		l.WriteToUDP([]byte(`{"IsBusy":false,"MachineType":"PC"}`), add)
	}
}

func (r relay) listenTCP(ctx context.Context) error {
	tcl, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero, Port: proto.TCPPort})
	if err != nil {
		return fmt.Errorf("error listening on tcp: %w", err)
	}
	for {
		c, err := tcl.Accept()
		if err != nil {
			log.Println("tcp err", err)
		} else {
			go func(c net.Conn) {
				addr := fmt.Sprintf("%s:%d", r.srv.Address, proto.TCPPort)
				server, err := net.Dial("tcp", addr)
				if err != nil {
					log.Fatal(err)
				}

				go hexSpy(c, server, "To Client")
				go hexSpy(server, c, "To Server")
				select {}
			}(c)
		}
	}
}

// Relay listens for clients and connects them to the local server.
func Relay(ctx context.Context) error {
	// TODO opts
	srv, err := proto.Discover(ctx)
	if err != nil {
		return fmt.Errorf("error discovering server: %w", err)
	}
	if len(srv) == 0 {
		return ErrorNoServer
	}
	r := relay{srv: &srv[0]}

	var meg multierrgroup.Group
	meg.GoWithContext(ctx, r.listenUDP)
	meg.GoWithContext(ctx, r.listenTCP)
	return meg.Wait()
}

func hexSpy(w io.Writer, r io.Reader, pre string) {
	for {
		p, err := proto.ReadPacket(r)
		if err != nil {
			log.Fatal(err)
		}
		switch p.PacketType {
		case proto.PacketTypeKeepAlive:
		case proto.PacketTypeDataUpdate:
			// go func() {
			// 	f, err := os.OpenFile("dataupdate.bin", os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
			// 	if err == nil {
			// 		fmt.Println(len(p.Body))
			// 		n, err := p.WriteTo(f)
			// 		if err != nil {
			// 			log.Println(err)
			// 		}
			// 		log.Println("wrote", n, "bytes")
			// 		err = f.Close()
			// 		if err != nil {
			// 			log.Println(err)
			// 		}
			// 	}
			// 	de, err := proto.UnmarshalDataEntries(p.Body)
			// 	if err != nil {
			// 		log.Println(err)
			// 	}
			// 	for _, e := range de {
			// 		fmt.Println(e.Name)
			// 	}
			// }()
		default:
			log.Printf("%s - Packet Type %s\n%s\n", pre, p.PacketType.String(), hex.Dump(p.Body))
		}
		p.WriteTo(w)
	}
}
