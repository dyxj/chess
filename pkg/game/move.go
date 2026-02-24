package game

import (
	"fmt"

	"github.com/dyxj/chess/pkg/engine"
)

// Move
// ( 56) ( 57) ( 58) ( 59) ( 60) ( 61) ( 62) ( 63)
// ( 48) ( 49) ( 50) ( 51) ( 52) ( 53) ( 54) ( 55)
// ( 40) ( 41) ( 42) ( 43) ( 44) ( 45) ( 46) ( 47)
// ( 32) ( 33) ( 34) ( 35) ( 36) ( 37) ( 38) ( 39)
// ( 24) ( 25) ( 26) ( 27) ( 28) ( 29) ( 30) ( 31)
// ( 16) ( 17) ( 18) ( 19) ( 20) ( 21) ( 22) ( 23)
// (  8) (  9) ( 10) ( 11) ( 12) ( 13) ( 14) ( 15)
// (  0) (  1) (  2) (  3) (  4) (  5) (  6) (  7)
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

type MoveResult struct {
	Color  engine.Color  `json:"color"`
	Symbol engine.Symbol `json:"symbol"`
	From   int           `json:"from"`
	To     int           `json:"to"`

	IsCastling bool `json:"isCastling"`
	RookFrom   int  `json:"rookFrom"`
	RookTo     int  `json:"rookTo"`

	Captured    engine.Symbol `json:"captured"`
	Promotion   engine.Symbol `json:"promotion"`
	IsEnPassant bool          `json:"isEnPassant"`
}

func fromEngine(m engine.Move) MoveResult {
	from := engine.MailboxToIndex(m.From)
	if from < 0 {
		panic(fmt.Sprintf("invalid 'from' index: %d", from))
	}
	to := engine.MailboxToIndex(m.To)
	if to < 0 {
		panic(fmt.Sprintf("invalid 'to' index: %d", to))
	}
	return MoveResult{
		Color:       m.Color,
		Symbol:      m.Symbol,
		From:        from,
		To:          to,
		IsCastling:  m.IsCastling,
		RookFrom:    m.RookFrom,
		RookTo:      m.RookTo,
		Captured:    m.Captured,
		Promotion:   m.Promotion,
		IsEnPassant: m.IsEnPassant,
	}
}
