package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/andrewstuart/gopip/command"
	"github.com/andrewstuart/gopip/pipdb"
	"github.com/andrewstuart/gopip/proto"
)

// Client listens and relays traffic
type Client struct {
	command.Commander
	tCli, tSrv *net.TCPConn
	cli        net.Addr
	cliAddr    string
	db         pipdb.Database
}

// Connect receives a server and connects to it
func (c *Client) Connect(s proto.Server) error {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", s.Address, proto.TCPPort))
	if err != nil {
		return err
	}

	c.Commander.W = conn

	var dbPrinted bool

	var lastP, p *proto.Packet
	for {
		p, err = proto.ReadPacket(conn)
		if err != nil {
			log.Println(err)
			if err == io.EOF {
				log.Println("Connection closed by server. Last packet follows")
				log.Println(hex.Dump(lastP.Body))
				// defer c.Connect(s)
			}
			break
		}
		lastP = p

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
				inv := pipdb.Inventory{I: myInventory, V: weightedValue}
				sort.Sort(&inv)

				tw := tabwriter.NewWriter(os.Stdout, 2, 2, 3, ' ', 0)

				for _, item := range inv.I {
					fmt.Fprintf(tw, "%s\t%f\n", item.Name, weightedValue(item))
				}

				tw.Flush()
				dbPrinted = true

				d := inv.I[0]
				log.Println(c.Execute(command.DropItem, d.HandleID, 1, getItem(c, 0, "Inventory", "Version"), d.StackID))
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

func weightedValue(i pipdb.InventoryItem) float32 {
	return i.Info.Value / i.Info.Weight
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

	bs, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return
	}

	fmt.Println(string(bs))
}
