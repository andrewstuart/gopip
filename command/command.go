package command

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/andrewstuart/gopip/proto"
)

//Type is the type for the command.type
type Type int

//Well-known commands
const (
	UseItem = Type(iota)
	DropItem
	SetFavorite
	ToggleComponentFavorite
	SortInventory
	ToggleQuest
	SetCustomMapMarker
	RemoveCustomMapMarker
	CheckFastTravel
	FastTravel
	MoveLocalMap
	ZoomLocalMap
	ToggleRadioStation
	RequestLocalMapSnapshot
	ClearIdle
)

//Command is the type for a Pip Boy fallout 4 command
type Command struct {
	Type Type          `json:"type"`
	Args []interface{} `json:"args"`
	ID   int           `json:"id"`
}

//Commander is an abstraction for writing commands
type Commander struct {
	W  io.ReadWriter
	id int
}

//Execute executes a command
func (c *Commander) Execute(cmd Type, args ...interface{}) (*Result, error) {
	if args == nil {
		args = make([]interface{}, 0)
	}
	if c.id == 0 {
		c.id++
	}

	if c.W == nil {
		return nil, io.ErrClosedPipe
	}

	command := Command{
		Type: cmd,
		Args: args,
		ID:   c.id,
	}

	j, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	p := proto.Packet{
		PacketType: proto.CommandPacket,
		Body:       j,
	}
	p.PacketType = proto.CommandPacket

	_, err = p.WriteTo(c.W)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Packet type %d\n%s", p.PacketType, hex.Dump(p.Body))

	c.id++
	r := Result{}

	// p, err := proto.ReadPacket(c.W)
	// if err != nil {
	// 	return nil, err
	// }

	// err = json.Unmarshal(p.Body, &r)
	// if err != nil {
	// 	return nil, err
	// }

	// if r.Success != true {
	// 	return &r, fmt.Errorf("unsuccessful response from server")
	// }

	return &r, nil
}

//Result is the response for a command
type Result struct {
	Allowed bool
	Success bool
	ID      int
	Message string
}
