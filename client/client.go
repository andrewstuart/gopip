package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
)

//Client listens and relays traffic
type Client struct {
	u          *net.UDPConn
	tCli, tSrv *net.TCPConn
	cli        net.Addr
	cliAddr    string
	db         proto.Database
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

	var dbPrinted bool

	var p *Packet
	for {
		p, err = ReadPacket(conn)
		if err != nil {
			if err == io.EOF {
				defer c.Connect(s)
				break
			}
		}

		log.Println(p.PacketType, p.Length)

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

			myInventory := make([]InventoryItem, 0, 10)

			if !dbPrinted {
				for _, list := range getItem(c, 0, "Inventory").(map[string]interface{}) {
					bs, err := json.Marshal(list)
					if err != nil {
						continue
					}

					var inv []InventoryItem
					err = json.Unmarshal(bs, &inv)
					if err != nil {
						continue
					}

					myInventory = append(myInventory, inv...)

				}
				myI := Inventory(myInventory)
				sort.Sort(&myI)
				for _, item := range myI {
					fmt.Printf("%s\t%f\n", item.Name, item.Info.Value/item.Info.Weight)
				}

				dbPrinted = true
			} else {
				for _, d := range des {
					if d.Type == ModifyEntry {
						printJson(c, d.ID)
					}
				}
			}
		}
	}
	return nil
}

func getItem(c *Client, base uint32, props ...string) interface{} {
	v := c.db.ToJSON(base)

	for _, p := range props {
		switch v1 := v.(type) {
		case map[string]interface{}:
			if v1[p] == nil {
				break
			}
			v = v1[p]
		case []interface{}:
			if i, err := strconv.Atoi(p); err == nil && i < len(v1) {
				v = v1[i]
				continue
			}
			break
		}
	}

	return v
}

func printJson(c *Client, i uint32, props ...string) {
	v := getItem(c, i, props...)

	bs, err := json.Marshal(v)
	if err != nil {
		return
	}

	fmt.Println(string(bs))
}
