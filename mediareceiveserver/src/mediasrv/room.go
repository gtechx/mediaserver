package mediasrv

import (
	"fmt"
)

type Room struct {
	id      string
	clients map[string]*client
}

func NewRoom(id string) *Room {
	fmt.Println("new room ", id)
	return &Room{id, make(map[string]*client)}
}
