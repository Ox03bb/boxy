package runtime

import (
	"errors"
	"sync"

	"github.com/Ox03bb/boxy/internal/box"
)

// Runtime holds runtime state for running boxes in memory.
// It stores a mapping from box ID -> *box.Box. The PTY file handle is stored
// on the `box.Box` itself (as `Box.Pty`). This package deliberately does not
// implement any PTY reuse logic — it only stores the Box (and thereby its
// PTY file handle if present).
type Runtime struct {
	mu    sync.RWMutex
	boxes map[string]*box.Box
}

func New() *Runtime {
	return &Runtime{
		boxes: make(map[string]*box.Box),
	}
}

func (r *Runtime) Add(b *box.Box) error {
	if b == nil {
		return errors.New("box is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.boxes[b.ID] = b
	return nil
}

func (r *Runtime) Get(id string) (*box.Box, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.boxes[id]
	if !ok {
		return nil, false
	}
	return b, true
}

func (r *Runtime) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.boxes, id)
}

func (r *Runtime) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.boxes))
	for id := range r.boxes {
		ids = append(ids, id)
	}
	return ids
}

func (r *Runtime) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.boxes)
}
