package app

import "time"

// Clock provides the current time. It is injected to make Service tests
// deterministic (e.g., to fix today's date when testing the Today engine).
type Clock interface {
	Now() time.Time
}

// SystemClock returns the local system time on every call.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().Local() }
