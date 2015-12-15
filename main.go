package main

import (
	"log"
	"net"
	"os"

	"github.com/andrewstuart/gopip/client"
	"github.com/andrewstuart/gopip/proto"
)

func main() {
	c := client.Client{}

	if len(os.Args) > 1 {
		ip := net.ParseIP(os.Args[1])
		err := c.Connect(proto.Server{IP: ip})
		if err != nil {
			log.Fatal(err)
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
