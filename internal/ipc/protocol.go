package ipc

import (
	"encoding/json"

	"github.com/Ox03bb/boxy/internal/box"
)

type Cmd string

const (
	RunC  Cmd = "run"
	StopC Cmd = "stop"
)

// Base Command
type Command struct {
	Cmd  Cmd    `json:"cmd"`
	Args CmdArg `json:"args,omitempty"`
}

type CmdArg interface {
	cmdarg()
}

// Run Command
type Run struct {
	Image box.Image `json:"image"`
	Name  string    `json:"name,omitempty"`
}

func (Run) cmdarg() {}

func (c *Command) UnmarshalJSON(data []byte) error {
	var aux struct {
		Cmd  string          `json:"cmd"`
		Args json.RawMessage `json:"args"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.Cmd = Cmd(aux.Cmd)

	switch c.Cmd {
	case RunC:
		var r Run
		if len(aux.Args) != 0 {
			if err := json.Unmarshal(aux.Args, &r); err != nil {
				return err
			}
		}
		// store pointer so the original code that passes &Run works with it
		c.Args = &r
	default:
		c.Args = nil
	}

	return nil
}
