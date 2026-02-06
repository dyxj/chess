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
// ---------------------------
// 8 | ♜| ♞| ♝| ♛| ♚| ♝| ♞| ♜|
// ---------------------------
// 7 | ♟| ♟| ♟| ♟| ♟| ♟| ♟| ♟|
// ---------------------------
// 6 | ·| ·| ·| ·| ·| ·| ·| ·|
// ---------------------------
// 5 | ·| ·| ·| ·| ·| ·| ·| ·|
// ---------------------------
// 4 | ·| ·| ·| ·| ·| ·| ·| ·|
// ---------------------------
// 3 | ·| ·| ·| ·| ·| ·| ·| ·|
// ---------------------------
// 2 |-♙|-♙|-♙|-♙|-♙|-♙|-♙|-♙|
// ---------------------------
// 1 |-♖|-♘|-♗|-♕|-♔|-♗|-♘|-♖|
// ---------------------------
// -   a  b  c  d  e  f  g  h
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
		return " ♚"
	case -5:
		return " ♛"
	case -4:
		return " ♜"
	case -3:
		return " ♝"
	case -2:
		return " ♞"
	case -1:
		return " ♟"
	case 0:
		return " ·"
	case 1:
		return "-♙"
	case 2:
		return "-♘"
	case 3:
		return "-♗"
	case 4:
		return "-♖"
	case 5:
		return "-♕"
	case 6:
		return "-♔"
	default:
		return " ?"
	}
}
