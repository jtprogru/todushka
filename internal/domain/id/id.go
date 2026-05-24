package id

import (
	"crypto/rand"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// ID is a stringified ULID encoded as Crockford base32 (26 chars).
type ID string

const ulidLen = 26

// ShortLen is the length of the human-readable short form used in CLI.
const ShortLen = 6

var (
	ErrEmpty   = errors.New("id: empty")
	ErrInvalid = errors.New("id: invalid format")
)

var (
	entropyMu sync.Mutex
	entropy   = ulid.Monotonic(rand.Reader, 0)
)

// New returns a freshly generated ULID-based ID.
func New() ID {
	entropyMu.Lock()
	defer entropyMu.Unlock()
	u := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return ID(u.String())
}

// crockfordAlphabet is the Crockford base32 alphabet used by ULID.
// I, L, O, U are excluded by spec.
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// Parse validates s as a ULID and returns it as an ID.
func Parse(s string) (ID, error) {
	if s == "" {
		return "", ErrEmpty
	}
	if len(s) != ulidLen {
		return "", ErrInvalid
	}
	upper := strings.ToUpper(s)
	for i := 0; i < len(upper); i++ {
		if !strings.ContainsRune(crockfordAlphabet, rune(upper[i])) {
			return "", ErrInvalid
		}
	}
	if _, err := ulid.Parse(upper); err != nil {
		return "", ErrInvalid
	}
	return ID(upper), nil
}

// shortOffset chooses ShortLen characters from inside the ULID's random
// portion. The first 10 chars of a ULID encode the millisecond timestamp;
// two IDs minted in the same ms would share that prefix, so we sample from
// the random tail starting at offset 10.
const shortOffset = 10

// Short returns a user-facing reference of ShortLen characters drawn from
// the random portion of the ULID. Same-millisecond collisions in the
// timestamp prefix are avoided this way.
func Short(i ID) string {
	s := strings.ToUpper(string(i))
	if len(s) < shortOffset+ShortLen {
		if len(s) < ShortLen {
			return s
		}
		return s[:ShortLen]
	}
	return s[shortOffset : shortOffset+ShortLen]
}
