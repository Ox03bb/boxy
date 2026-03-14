package ipc

import (
	"fmt"
)

func Clinet() {
	var msg string
	c, err := Connect(DefaultSocketPath)

	if err != nil {
		panic("Conn error")
	}

	for {
		fmt.Scan(&msg)
		msg += "\n"
		Send(c, msg)
		if err != nil {
			println(err.Error())
		}
		Recive(c)
	}
}
