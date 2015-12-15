package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"

	"github.com/andrewstuart/gopip/pipdb"
	"github.com/andrewstuart/gopip/proto"
)

//Client listens and relays traffic
type Client struct {
	tCli, tSrv *net.TCPConn
	cli        net.Addr
	cliAddr    string
	db         pipdb.Database
}

//Connect receives a server and connects to it
func (c *Client) Connect(s proto.Server) error {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", s.Address, proto.TCPPort))
	if err != nil {
		return err
	}

	var dbPrinted bool

	var p *proto.Packet
	for {
		p, err = proto.ReadPacket(conn)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by server. Last packet follows")
				log.Println(hex.Dump(p.Body))
				defer c.Connect(s)
				break
			}
		}

		switch p.PacketType {
		case proto.KeepAlivePacket:
			p.WriteTo(conn)
		case proto.DataUpdatePacket:
			des, err := proto.UnmarshalDataEntries(p.Body)
			if err != nil {
				log.Println(err)
				continue
			}

			c.db.Update(des)

			myInventory := make([]pipdb.InventoryItem, 0, 10)

			if !dbPrinted {
				for _, list := range getItem(c, 0, "Inventory").(map[string]interface{}) {
					bs, err := json.Marshal(list)
					if err != nil {
						continue
					}

					var inv []pipdb.InventoryItem
					err = json.Unmarshal(bs, &inv)
					if err != nil {
						continue
					}

					myInventory = append(myInventory, inv...)

				}
				myI := pipdb.Inventory(myInventory)
				sort.Sort(&myI)
				for _, item := range myI {
					fmt.Printf("%s\t%f\n", item.Name, item.Info.Value/item.Info.Weight)
				}

				dbPrinted = true
			} else {
				for _, d := range des {
					if d.Type == proto.ModifyEntry {
						printJSON(c, d.ID)
					}
				}
			}
		}
	}
	return nil
}

func getItem(c *Client, base uint32, props ...string) interface{} {
	v := c.db.ToTree(base)

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

func printJSON(c *Client, i uint32, props ...string) {
	v := getItem(c, i, props...)

	bs, err := json.Marshal(v)
	if err != nil {
		return
	}

	fmt.Println(string(bs))
}
