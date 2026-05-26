package main

import (
	"testing"
)

// TestHasReadOnlyFlag verifies that hasReadOnlyFlag recognizes every
// standard cobra/pflag boolean form for both --readonly and the --ro
// alias, including the --flag=value form. This guards against a class
// of bugs where the pre-scan diverges from cobra's parser and a user
// believes "--readonly=true" forces read-only mode but it does not
// (regression source — see review.md F-1).
func TestHasReadOnlyFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", []string{}, false},
		{"bare --readonly", []string{"--readonly"}, true},
		{"bare --ro", []string{"--ro"}, true},
		{"--readonly=true", []string{"--readonly=true"}, true},
		{"--readonly=1", []string{"--readonly=1"}, true},
		{"--readonly=t", []string{"--readonly=t"}, true},
		{"--readonly=TRUE", []string{"--readonly=TRUE"}, true},
		{"--ro=true", []string{"--ro=true"}, true},
		{"--readonly=false", []string{"--readonly=false"}, false},
		{"--readonly=0", []string{"--readonly=0"}, false},
		{"--ro=false", []string{"--ro=false"}, false},
		{"after --config string flag", []string{"--config", "/tmp/cfg", "--readonly"}, true},
		{"--readonly=true with --config later", []string{"--readonly=true", "--config", "/tmp/cfg"}, true},
		{"--config=path --readonly=true", []string{"--config=path", "--readonly=true"}, true},
		// Persistent flags in cobra apply wherever they appear (before
		// or after the subcommand). The pre-scan mirrors that and does
		// NOT stop at a positional arg — otherwise "complete --readonly"
		// would diverge from cobra's parsed view.
		{"--readonly after subcommand", []string{"complete", "--readonly"}, true},
		{"--readonly=true after subcommand", []string{"complete", "--readonly=true"}, true},
		{"unrelated flag only", []string{"--config", "/tmp/cfg"}, false},
		{"garbled value treated as false", []string{"--readonly=notabool"}, false},
		{"--readonly= (empty value)", []string{"--readonly="}, true},
		{"--ro= (empty value)", []string{"--ro="}, true},
		{"only positional args", []string{"complete", "task-id"}, false},
		{"--ro=false anywhere", []string{"complete", "--ro=false"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasReadOnlyFlag(tc.args)
			if got != tc.want {
				t.Fatalf("hasReadOnlyFlag(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}
