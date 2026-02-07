package room

import (
	"math/rand"
	"strings"
)

const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode() string {
	sb := strings.Builder{}
	for i := 0; i < 6; i++ {
		// validChars contains only ASCII characters(single byte chars)
		// so byte indexing is safe
		sb.WriteByte(validChars[rand.Intn(len(validChars))])
	}
	return sb.String()
}
