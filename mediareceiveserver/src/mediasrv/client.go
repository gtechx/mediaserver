package mediasrv

import (
	"fmt"
	"io"
)

type client struct {
	id  string
	rwc io.ReadWriteCloser
}

func newClient(id string) *client {
	c := client{id: id}
	fmt.Println("new client ", id)
	return &c
}
