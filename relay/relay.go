package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
)

var serverIP = &net.UDPAddr{IP: net.IP([]byte{192, 168, 16, 12}), Port: UDPPort}

//Relay listens for clients and connects them to the local server.
func (r *Client) Relay() error {
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

	log.Println(r.Discover())

	l := r.u

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
			l.WriteToUDP(bs[:n], serverIP)
		default:
			if r.cli == nil {
				r.cli = addr
				r.cliAddr = addr.String()
				l.WriteToUDP(bs[:n], serverIP)
			}
			fmt.Printf("r = %+v\n", r)
		}

		if addr.String() != "192.168.16.15:28000" {
			if m["cmd"] != nil || m["MachineType"] != nil {
			}
		}
	}
}

var db = make(map[uint32]*DataEntry, 65000)

func hexSpy(w io.Writer, r io.Reader, pre string) {
	for {
		p, err := ReadPacket(r)
		if err != nil {
			log.Fatal(err)
		}
		p.WriteTo(w)

		if p.PacketType == DataUpdatePacket {

			for ct := 0; ct < len(p.Body); {
				d, n, err := UnmarshalDataEntry(p.Body[ct:])
				if err != nil {
					log.Println("Error unmarshalling", err)
					fmt.Printf("d = %+v\n", d)
					break
				}

				if d.Type == 8 {
					v := d.Value.(InsRemove)
					for _, ins := range v.Insert {
						db[ins.Ref].Name = ins.Name
					}
				}

				if _, ok := db[d.ID]; !ok {
					db[d.ID] = d
				} else {
					db[d.ID].Value = d.Value
				}

				if db[d.ID].Name == "TimeHour" {
					v := db[d.ID].Value.(float32)
					hComp := int(v)
					mComp := int((v - float32(hComp)) * 60)
					fmt.Printf("%d:%d\n", hComp, mComp)
				} else {
					log.Println(db[d.ID])
				}

				ct += n
			}
		}
	}
}
