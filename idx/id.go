package idx

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

func UUID() uuid.UUID {
	return uuid.New()
}

func ULID(monotonic ...bool) ulid.ULID {
	entropy := rand.Reader
	if len(monotonic) == 1 && monotonic[0] {
		entropy = ulid.Monotonic(entropy, 0)
	}
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
}
