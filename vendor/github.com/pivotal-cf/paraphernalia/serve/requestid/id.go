// Package requestid can be used to generate IDs that are suitable for use in
// request tracing and correlation.
package requestid

import (
	"math/rand"

	"github.com/oklog/ulid"
)

var entropy = rand.New(rand.NewSource(int64(ulid.Now())))

// Generate creates a new ID.
func Generate() string {
	return ulid.MustNew(ulid.Now(), entropy).String()
}
