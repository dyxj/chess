package faker

import (
	"math/rand/v2"

	"github.com/dyxj/chess/internal/engine"
)

func Color() engine.Color {
	return engine.Colors[rand.IntN(len(engine.Colors))]
}
