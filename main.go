package main

import "github.com/andrewstuart/pb/relay"

func main() {
	r := &relay.Relay{}

	r.Listen()
}
