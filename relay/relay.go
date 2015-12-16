package relay

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/andrewstuart/gopip/proto"
)

var serverIP = net.IP{192, 168, 16, 12}
var serverUDP = net.UDPAddr{IP: serverIP, Port: proto.UDPPort}

//Relay listens for clients and connects them to the local server.
func Relay() error {
	go func() {
		tcl, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero, Port: proto.TCPPort})
		if err != nil {
			log.Println("err", err)
		}

		for {
			c, err := tcl.Accept()
			if err != nil {
				log.Println("tcp err", err)
			} else {
				go func(c net.Conn) {
					server, err := net.Dial("tcp", "192.168.16.12:27000")
					if err != nil {
						log.Fatal(err)
					}

					go hexSpy(c, server, "To Client")
					go hexSpy(server, c, "To Server")
					select {}
				}(c)
			}
		}
	}()

	var cli net.Addr
	l, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: proto.UDPPort})
	if err != nil {
		return err
	}

	bs := make([]byte, 1024)
	for {
		n, addr, err := l.ReadFrom(bs)
		if err != nil {
			return err
		}

		m := make(map[string]interface{})
		json.Unmarshal(bs[:n], &m)
		fmt.Printf("m = %+v\n", m)

		if addr.String() == "192.168.16.15:28000" {
			continue
		}

		switch addr.String() {
		case "192.168.16.12:28000":
			if cli != nil {
				l.WriteToUDP(bs[:n], cli.(*net.UDPAddr))
			}
		case cli.String():
			l.WriteToUDP(bs[:n], &serverUDP)
		default:
			if cli == nil {
				cli = addr
				l.WriteToUDP(bs[:n], &serverUDP)
			}
		}

		if addr.String() != "192.168.16.15:28000" {
			if m["cmd"] != nil || m["MachineType"] != nil {
			}
		}
	}
}

func hexSpy(w io.Writer, r io.Reader, pre string) {
	for {
		p, err := proto.ReadPacket(r)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s - Packet Type %d\n%s\n", pre, p.PacketType, hex.Dump(p.Body))
		p.WriteTo(w)
	}
}
