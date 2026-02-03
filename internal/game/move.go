package game

import "github.com/dyxj/chess/internal/engine"

type Move struct {
	Color  engine.Color
	Symbol engine.Symbol
	From   int
	To     int
}

func (m Move) mbTo() int {
	return engine.IndexToMailbox(m.To)
}

func (m Move) mbFrom() int {
	return engine.IndexToMailbox(m.From)
}
