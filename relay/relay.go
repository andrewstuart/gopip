package relay

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

//Pip boy ports
const (
	UDPPort = 28000
	TCPPort = 27000
)

//Relay listens and relays traffic
type Relay struct {
	u, t    net.Conn
	cli     net.Addr
	cliAddr string
}

var serverIp = &net.UDPAddr{IP: net.IP([]byte{192, 168, 16, 12}), Port: UDPPort}

func (r *Relay) Listen() error {
	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: UDPPort}

	go func() {
		tcl, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero, Port: TCPPort})
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

	l, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return err
	}
	r.u = l

	_, err = l.WriteTo([]byte(`{"cmd": "autodiscover"}`), serverIp)
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
			if r.cli != nil {
				l.WriteToUDP(bs[:n], r.cli.(*net.UDPAddr))
			}
		case r.cliAddr:
			l.WriteToUDP(bs[:n], serverIp)
		default:
			if r.cli == nil {
				r.cli = addr
				r.cliAddr = addr.String()
				l.WriteToUDP(bs[:n], serverIp)
			}
			fmt.Printf("r = %+v\n", r)
		}

		if addr.String() != "192.168.16.15:28000" {
			if m["cmd"] != nil || m["MachineType"] != nil {
			}
		}
	}

	io.Copy(os.Stdout, l)

	return nil
}

func hexSpy(w io.Writer, r io.Reader, pre string) {
	for {
		p, err := ReadPacket(r)
		if err != nil {
			log.Fatal(err)
		}

		//Dump
		fmt.Fprintf(os.Stdout, "%s: channel %d, length %d\n", pre, p.Channel, p.Length)
		fmt.Fprintln(os.Stdout, hex.Dump(p.Body))

		p.WriteTo(w)
	}
}

type Packet struct {
	Channel      uint8
	Length       uint32
	Body, header []byte
}

func ReadPacket(r io.Reader) (*Packet, error) {
	p := &Packet{
		header: make([]byte, 5),
	}

	_, err := r.Read(p.header)
	if err != nil {
		return nil, err
	}

	p.Channel = p.header[4]
	err = binary.Read(bytes.NewReader(p.header[:4]), binary.LittleEndian, &p.Length)
	if err != nil {
		return nil, err
	}

	p.Body = make([]byte, p.Length)
	var tot uint32

	for tot < p.Length {
		n, err := r.Read(p.Body[tot:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		tot += uint32(n)
	}

	return p, nil
}

//WriteTo sends a packet to a writer
func (p *Packet) WriteTo(w io.Writer) error {
	_, err := w.Write(p.header)
	if err != nil {
		return err
	}
	_, err = w.Write(p.Body)
	return err
}
