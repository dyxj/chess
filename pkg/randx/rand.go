package randx

import "math/rand/v2"

func FromSlice[T any](slice []T) T {
	return slice[rand.IntN(len(slice))]
}

func Bool() bool {
	return FromSlice([]bool{true, false})
}
