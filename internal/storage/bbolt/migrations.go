package bbolt

import (
	bolt "go.etcd.io/bbolt"
)

// Migration applies an incremental schema change. It runs in its own
// transaction; the framework writes the new schema_version after Apply
// returns nil.
type Migration struct {
	Version int
	Apply   func(tx *bolt.Tx) error
}

// v1 is a no-op: ensureBuckets() in Open() already creates every bucket the
// initial schema needs. The migration exists to record the schema version so
// future binaries can detect upgrade direction.
var v1 = Migration{
	Version: 1,
	Apply:   func(_ *bolt.Tx) error { return nil },
}

var migrations = []Migration{v1}

func migrationByVersion(v int) (Migration, bool) {
	for _, m := range migrations {
		if m.Version == v {
			return m, true
		}
	}
	return Migration{}, false
}
