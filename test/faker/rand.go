package faker

import (
	"math/rand/v2"

	"github.com/dyxj/chess/internal/engine"
)

var colors = []engine.Color{engine.White, engine.Black}

func Color() engine.Color {
	return colors[rand.IntN(len(colors))]
}
