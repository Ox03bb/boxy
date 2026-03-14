package ipc

import "github.com/Ox03bb/boxy/internal/box"

type Cmd string

const (
	RunC  Cmd = "run"
	StopC Cmd = "stop"
)

// Base Command
type Command struct {
	Cmd  Cmd ``
	Args CmdArg
}

type CmdArg interface {
	cmdarg()
}

// Run Command

type Run struct {
	image box.Image `json:image`
	name  string    `json:name,omitempty`
}

func (Run) cmdarg() {}
