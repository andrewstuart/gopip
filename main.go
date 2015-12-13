package main

import (
	"log"

	"github.com/andrewstuart/gopip/relay"
)

func main() {
	r, err := relay.NewClient()
	if err != nil {
		log.Fatal(err)
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
