package cli

import (
	"fmt"
	"strings"
)

const fileHeader = "    a  b  c  d  e  f  g  h "
const lineBreak = "\n---------------------------\n"

var rankHeader = []string{"1", "2", "3", "4", "5", "6", "7", "8"}

func (a *Adapter) Render() string {
	sb := strings.Builder{}
	sb.Grow(476)

	sb.WriteString(fileHeader)
	sb.WriteString(lineBreak)

	b := a.g.GridRaw()
	x := 0
	fhIndex := 1
	sb.WriteString(rankHeader[0] + " |")
	for i := len(b) - 1; i >= 0; i-- {
		sb.WriteString(fmt.Sprintf("%2d|", b[i]))
		if x >= 7 && i > 0 {
			sb.WriteString(lineBreak)
			sb.WriteString(fmt.Sprintf("%s |", fileHeader[fhIndex]))
			x = 0
			fhIndex++
		} else {
			x++
		}
	}
	sb.WriteString("\n")
	return sb.String()
}
