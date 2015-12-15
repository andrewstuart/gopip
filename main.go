package main

import (
	"log"
	"net"
	"os"

	"github.com/andrewstuart/gopip/relay"
)

func main() {
	r, err := relay.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		ip := net.ParseIP(os.Args[1])
		err = r.Connect(relay.Server{IP: ip})
		if err != nil {
			log.Fatal(err)
		}
	}

	servers, err := r.Discover()
	if err != nil {
		log.Fatal(err)
	}

	if len(servers) < 1 {
		log.Fatal("No servers found")
	}

	err = r.Connect(servers[0])
	if err != nil {
		log.Fatal(err)
	}
}
