package ipc

import (
	"fmt"
	"net"
)

func clinet() {
	c, err := net.Dial("unix", SocketPath)
	var msg string
	if err != nil {
		panic(err.Error())
	}
	for {
		fmt.Scan(&msg)
		_, err := c.Write([]byte("msg\n"))
		if err != nil {
			println(err.Error())
		}

	}
}
