package engine

import "fmt"

type Color int

const (
	White           Color = 1
	Black           Color = -1
	colorWhiteStr         = "white"
	colorBlackStr         = "black"
	colorUnknownStr       = "unknown"
)

var Colors = []Color{White, Black}

func (c Color) Opposite() Color {
	return c * -1
}

func (c Color) String() string {
	switch c {
	case White:
		return colorWhiteStr
	case Black:
		return colorBlackStr
	default:
		return colorUnknownStr
	}
}

func (c Color) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

//goland:noinspection GoMixedReceiverTypes
func (c *Color) UnmarshalText(text []byte) error {
	str := string(text)
	switch str {
	case colorWhiteStr:
		*c = White
	case colorBlackStr:
		*c = Black
	default:
		return fmt.Errorf("unknown color: %s valid color(white,black)", str)
	}
	return nil
}
