package main

import (
	"log"
	"os"

	"github.com/andrewstuart/gopip/client"
	_ "github.com/andrewstuart/gopip/command"
	"github.com/andrewstuart/gopip/proto"
	"github.com/andrewstuart/gopip/relay"
	"github.com/gopuff/morecontext"
)

func main() {
	ctx := morecontext.ForSignals()
	c := client.Client{}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "relay":
			err := relay.Relay(ctx)
			if err != nil {
				log.Fatal(err)
			}
			return
		case "connect", "c", "conn":
			addr := "localhost"
			if len(os.Args) > 2 {
				addr = os.Args[2]
			}
			err := c.Connect(proto.Server{Address: addr})
			if err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	servers, err := proto.Discover(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(servers) < 1 {
		log.Fatal("No servers found")
	}

	log.Println("Discovered server at ", servers[0].Address)

	err = c.Connect(servers[0])
	if err != nil {
		log.Fatal(err)
	}
}
