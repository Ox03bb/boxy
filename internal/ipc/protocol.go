package ipc

import (
	"encoding/json"

	"github.com/Ox03bb/boxy/internal/box"
)

type Cmd string

const (
	RunC    Cmd = "run"
	AttachC Cmd = "attach"
	PsC     Cmd = "ps"
	StopC   Cmd = "stop"
	RmC     Cmd = "rm"
	ExecC   Cmd = "exec"
)

// ! ================= Base Command ==================
type Command struct {
	Cmd  Cmd    `json:"cmd"`
	Args CmdArg `json:"args,omitempty"`
}

type CmdArg interface {
	cmdarg()
}

// ================== Run Command ==================
type Run struct {
	Image box.Image `json:"image"`
	Name  string    `json:"name,omitempty"`
}

func (Run) cmdarg() {}

// ==================attach Command ==================
type Attach struct {
	BoxIdentifier string `json:"box_id"`
	Is_name       bool   `json:"is_name"`
}

func (Attach) cmdarg() {}

// ================== exec Command ==================
type Exec struct {
	BoxIdentifier string   `json:"box_id"`
	Is_name       bool     `json:"is_name"`
	Cmd           []string `json:"cmd,omitempty"`
	Tty           bool     `json:"tty,omitempty"`
	Interactive   bool     `json:"interactive,omitempty"`
}

func (Exec) cmdarg() {}

// ================== ps Command ==================
type Ps struct{}

func (Ps) cmdarg() {}

// ================== rm Command ==================
type Rm struct {
	BoxIdentifier string `json:"box_id"`
	Is_name       bool   `json:"is_name"`
}

func (Rm) cmdarg() {}

// ================== UnmarshalJSON for Command ==================

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
		c.Args = &r
	case AttachC:
		var a Attach
		if len(aux.Args) != 0 {
			if err := json.Unmarshal(aux.Args, &a); err != nil {
				return err
			}
		}
		c.Args = &a
	case ExecC:
		var e Exec
		if len(aux.Args) != 0 {
			if err := json.Unmarshal(aux.Args, &e); err != nil {
				return err
			}
		}
		c.Args = &e
	case RmC:
		var r Rm
		if len(aux.Args) != 0 {
			if err := json.Unmarshal(aux.Args, &r); err != nil {
				return err
			}
		}
		c.Args = &r
	case PsC:
		// ps has no args
		c.Args = &Ps{}
	default:
		c.Args = nil
	}

	return nil
}
