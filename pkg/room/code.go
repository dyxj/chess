package room

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	// deep dive required here, added lock for now
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
	rngMu sync.Mutex
)

func generateCode() string {
	rngMu.Lock()
	defer rngMu.Unlock()

	sb := strings.Builder{}

	for i := 0; i < 6; i++ {
		// validChars contains only ASCII characters(single byte chars)
		// so byte indexing is safe
		sb.WriteByte(validChars[rng.Intn(len(validChars))])
	}

	return sb.String()
}
