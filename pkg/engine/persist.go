package engine

import (
	"bytes"
	"encoding/gob"
	"io"
)

// Save serializes the board state using gob encoding.
func (b *Board) Save(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	return encoder.Encode(b)
}

// Load deserializes the board state using gob encoding.
func (b *Board) Load(r io.Reader) error {
	decoder := gob.NewDecoder(r)
	return decoder.Decode(b)
}

func (b *Board) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(boardData{
		Cells:                  b.cells,
		WhitePieces:            b.whitePieces,
		BlackPieces:            b.blackPieces,
		WhiteKingPos:           b.whiteKingPos,
		BlackKingPos:           b.blackKingPos,
		RoundHistory:           b.roundHistory,
		ActiveColor:            b.activeColor,
		Graveyard:              b.graveyard,
		DrawCounter:            b.drawCounter,
		BoardStateHashMapCount: b.boardStateHashMapCount,
	})
	return buf.Bytes(), err
}

func (b *Board) GobDecode(data []byte) error {
	var d boardData
	reader := bytes.NewReader(data)
	if err := gob.NewDecoder(reader).Decode(&d); err != nil {
		return err
	}
	b.cells = d.Cells
	b.whitePieces = setCapIfNil(d.WhitePieces, 16)
	b.blackPieces = setCapIfNil(d.BlackPieces, 16)
	b.whiteKingPos = d.WhiteKingPos
	b.blackKingPos = d.BlackKingPos
	b.roundHistory = setCapIfNil(d.RoundHistory, 256)
	b.activeColor = d.ActiveColor
	b.graveyard = setCapIfNil(d.Graveyard, 32)
	b.drawCounter = d.DrawCounter
	b.boardStateHashMapCount = setMapCapIfNil(d.BoardStateHashMapCount, 256)
	return nil
}

func (p *Piece) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(pieceData{
		Symbol:    p.symbol,
		Color:     p.color,
		Position:  p.position,
		MoveCount: p.moveCount,
	})
	return buf.Bytes(), err
}

func (p *Piece) GobDecode(data []byte) error {
	var d pieceData
	reader := bytes.NewReader(data)
	if err := gob.NewDecoder(reader).Decode(&d); err != nil {
		return err
	}
	p.symbol = d.Symbol
	p.color = d.Color
	p.position = d.Position
	p.moveCount = d.MoveCount
	return nil
}

type boardData struct {
	Cells                  [boardSize]int
	WhitePieces            []Piece
	BlackPieces            []Piece
	WhiteKingPos           int
	BlackKingPos           int
	RoundHistory           []round
	ActiveColor            Color
	Graveyard              []Piece
	DrawCounter            int
	BoardStateHashMapCount map[uint64]int
}

type pieceData struct {
	Symbol    Symbol
	Color     Color
	Position  int
	MoveCount int
}

func setCapIfNil[T any](slice []T, capacity int) []T {
	if slice == nil {
		slice = make([]T, 0, capacity)
	}
	return slice
}

func setMapCapIfNil[K comparable, V any](m map[K]V, capacity int) map[K]V {
	if m == nil {
		m = make(map[K]V, capacity)
	}
	return m
}
