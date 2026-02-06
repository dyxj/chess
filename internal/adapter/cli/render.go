package cli

import (
	"fmt"
	"strconv"
	"strings"
)

const fileHeader = "    a  b  c  d  e  f  g  h "
const lineBreak = "\n---------------------------\n"

var rankHeader = []string{"8", "7", "6", "5", "4", "3", "2", "1"}

// Render returns board representation
//
//	---------------------------
//	8 |-♖|-♘|-♗|-♕|-♔|-♗|-♘|-♖|
//	---------------------------
//	7 |-♙|-♙|-♙|-♙|-♙|-♙|-♙|-♙|
//	---------------------------
//	6 | ·| ·| ·| ·| ·| ·| ·| ·|
//	---------------------------
//	5 | ·| ·| ·| ·| ·| ·| ·| ·|
//	---------------------------
//	4 | ·| ·| ·| ·| ·| ·| ·| ·|
//	---------------------------
//	3 | ·| ·| ·| ·| ·| ·| ·| ·|
//	---------------------------
//	2 | ♟| ♟| ♟| ♟| ♟| ♟| ♟| ♟|
//	---------------------------
//	1 | ♜| ♞| ♝| ♛| ♚| ♝| ♞| ♜|
//	---------------------------
//		a  b  c  d  e  f  g  h
func (a *Adapter) Render() string {
	sb := strings.Builder{}
	sb.Grow(600)

	sb.WriteString(lineBreak)
	b := a.g.GridRaw()
	fhIndex := 1
	sb.WriteString(rankHeader[0] + " |")
	for x := 7; x >= 0; x-- {
		for y := 0; y < 8; y++ {
			i := x*8 + y
			sb.WriteString(fmt.Sprintf("%2s|", a.iconMapper(b[i])))
		}
		sb.WriteString(lineBreak)
		if fhIndex < 8 {
			sb.WriteString(fmt.Sprintf("%s |", rankHeader[fhIndex]))
			fhIndex++
		}
	}
	sb.WriteString(fileHeader)
	return sb.String()
}

type iconMapper func(int) string

func numberIconMapper(pValue int) string {
	return strconv.Itoa(pValue)
}

func symbolIconMapper(pValue int) string {
	switch pValue {
	case -6:
		return "-♔" // black king
	case -5:
		return "-♕" // black queen
	case -4:
		return "-♖" // black rook
	case -3:
		return "-♗" // black bishop
	case -2:
		return "-♘" // black knight
	case -1:
		return "-♙" // black pawn
	case 0:
		return " ·" // empty
	case 1:
		return " ♟" // white pawn
	case 2:
		return " ♞" // white knight
	case 3:
		return " ♝" // white bishop
	case 4:
		return " ♜" // white rook
	case 5:
		return " ♛" // white queen
	case 6:
		return " ♚" // white king
	default:
		return " ?"
	}
}
