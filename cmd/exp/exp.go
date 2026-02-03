package main

import (
	"fmt"

	engine2 "github.com/dyxj/chess/internal/engine"
)

// Board dimensions
const boardWidth = 10  // column
const boardHeight = 12 // row

func main() {
	PrintBoard()
	//ShowDraftBoard()
	//fmt.Println(engine2.Color(0) == engine2.White)
	//fmt.Println(engine2.Color(0) == engine2.Black)
}

func PrintBoard() {
	board := engine2.NewBoard()
	fmt.Println(board.GridFull())
	fmt.Println(board.Grid())
	fmt.Println(board.GridRaw())
}

func ShowDraftBoard() {
	// Start from bottom row (index 0) and go up to top row (index 119)
	for x := boardHeight - 1; x >= 0; x-- {
		for y := 0; y < boardWidth; y++ {
			i := x*boardWidth + y
			fmt.Printf("(%3d) ", i)
		}
		fmt.Println()
	}

	fmt.Println()

	i := 0
	for x := 0; x < boardHeight; x++ {
		for y := 0; y < boardWidth; y++ {
			fmt.Printf("(%3d) ", i)
			i++
		}
		fmt.Println()
	}
}
