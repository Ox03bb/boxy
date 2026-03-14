package ipc

import (
	"fmt"
	"net"
	"os"
)

func echoServer(c net.Conn) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]
		fmt.Printf("Received: %v", string(data))
		_, err = c.Write(data)
		if err != nil {
			panic("Write: " + err.Error())
		}
	}
}

func Server() {
	os.Remove(DefaultSocketPath)

	l, err := net.Listen("unix", DefaultSocketPath)
	if err != nil {
		println("listen error", err.Error())
		return
	}

	defer os.Remove(DefaultSocketPath)

	for {
		fd, err := l.Accept()
		if err != nil {
			println("accept error", err.Error())
			return
		}

		go echoServer(fd)
	}
}
