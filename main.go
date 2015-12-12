package main

import "github.com/andrewstuart/gopip/relay"

func main() {
	r := &relay.Relay{}
	r.Listen()
}
