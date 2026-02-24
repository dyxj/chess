package room

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeGenerator(t *testing.T) {
	code := generateCode()

	assert.Len(t, code, 6)

	for _, char := range code {
		assert.True(t, (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'))
	}
}
