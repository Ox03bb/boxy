package cli

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Ox03bb/boxy/internal/ipc"
)

func client(req *ipc.Command, sock *net.Conn) error {

	defer ipc.Close(*sock)

	msg, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	if err := ipc.Send(*sock, msg); err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	response, err := ipc.Recive(*sock)
	if err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	fmt.Println("Response:", string(response))
	return nil
}
