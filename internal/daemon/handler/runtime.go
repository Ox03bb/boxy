package handler

import "github.com/Ox03bb/boxy/internal/runtime"

// rt holds the runtime instance used by daemon handlers.
var rt *runtime.Runtime

// SetRuntime sets the runtime instance for handlers to use.
func SetRuntime(r *runtime.Runtime) {
	rt = r
}

// Runtime returns the currently set runtime (may be nil).
func Runtime() *runtime.Runtime {
	return rt
}
