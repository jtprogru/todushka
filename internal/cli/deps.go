// Package cli provides the Cobra command surface for non-interactive use
// and the entry point that launches the TUI.
package cli

import (
	"io"
	"os"

	"github.com/jtprogru/todushka/internal/app"
)

// Deps is the test seam: callers wire real os streams in production and
// in-memory buffers in tests.
type Deps struct {
	Service   *app.Service
	Stdout    io.Writer
	Stderr    io.Writer
	Stdin     io.Reader
	Env       func(string) string
	LaunchTUI func(*app.Service) error // returns an error to surface to main
}

func DefaultDeps(svc *app.Service) Deps {
	return Deps{
		Service: svc,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
		Env:     os.Getenv,
	}
}
