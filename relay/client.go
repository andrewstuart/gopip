package relay

import (
	"fmt"
	"log"
	"net"
)

//Client listens and relays traffic
type Client struct {
	u          *net.UDPConn
	tCli, tSrv *net.TCPConn
	cli        net.Addr
	db         Database
	cliAddr    string
}

func NewClient() (*Client, error) {
	c := Client{}
	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: UDPPort}

	l, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return nil, err
	}
	c.u = l

	return &c, nil
}

//PacketType accounts for the possible packet types
type PacketType uint8

//Well-known packet types
const (
	KeepAlivePacket = PacketType(iota)
	ConnecctionAcceptedPacket
	ConnectionRefusedPacket
	DataUpdatePacket
	MapUpdatePacket
	CommandPacket
)

func (c *Client) Connect(s Server) error {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", s.IP, TCPPort))
	if err != nil {
		return err
	}

	var p *Packet
	for {
		p, err = ReadPacket(conn)
		switch p.PacketType {
		case KeepAlivePacket:
			p.WriteTo(conn)
		case DataUpdatePacket:
			des, err := UnmarshalDataEntries(p.Body)
			if err != nil {
				log.Println(err)
				continue
			}

			c.db.Update(des)

			for _, d := range des {
				fmt.Printf("d = %+v\n", d)
				if d.Children != nil {
					for _, child := range d.Children {
						fmt.Printf("child = %+v\n", child)
					}
				}
			}
		}
	}
}
