package cli

import (
	"encoding/json"
	"fmt"

	"github.com/Ox03bb/boxy/internal/ipc"
)

func Client(req *ipc.Command) error {
	c, err := ipc.Connect("")
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer ipc.Close(c)

	msg, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	if err := ipc.Send(c, msg); err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	response, err := ipc.Recive(c)
	if err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	fmt.Println("Response:", string(response))
	return nil
}
