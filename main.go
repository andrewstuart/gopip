package main

import (
	"log"
	"os"

	"github.com/andrewstuart/gopip/client"
	_ "github.com/andrewstuart/gopip/command"
	"github.com/andrewstuart/gopip/proto"
	"github.com/andrewstuart/gopip/relay"
)

func main() {
	c := client.Client{}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "relay":
			err := relay.Relay()
			if err != nil {
				log.Fatal(err)
			}
			return
		case "connect", "c", "conn":
			if len(os.Args) > 2 {
				err := c.Connect(proto.Server{Address: os.Args[2]})
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		servers, err := proto.Discover()
		if err != nil {
			log.Fatal(err)
		}

		if len(servers) < 1 {
			log.Fatal("No servers found")
		}

		err = c.Connect(servers[0])
		if err != nil {
			log.Fatal(err)
		}
	}
}
