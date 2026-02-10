package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/game"
	"github.com/google/uuid"
)

type Player struct {
	ID   uuid.UUID
	Name string
}

type Adapter struct {
	whitePlayer Player
	blackPlayer Player
	g           *game.Game
	reader      *bufio.Scanner
	writer      *bufio.Writer
	iconMapper  iconMapper
}

type Option func(*Adapter)

func WithNumberIconMapper() Option {
	return func(a *Adapter) {
		a.iconMapper = numberIconMapper
	}
}

func WithSymbolIconMapper() Option {
	return func(a *Adapter) {
		a.iconMapper = symbolIconMapper
	}
}

func NewAdapter(option ...Option) *Adapter {
	a := &Adapter{
		reader:     bufio.NewScanner(os.Stdin),
		writer:     bufio.NewWriter(os.Stdout),
		iconMapper: symbolIconMapper,
	}

	for _, opt := range option {
		opt(a)
	}

	return a
}

func (a *Adapter) Run() {
	a.writeIntro()
	wName := a.requestForPlayerName(engine.White)
	bName := a.requestForPlayerName(engine.Black)
	a.setWhitePlayer(Player{
		ID:   uuid.New(),
		Name: wName,
	})
	a.setBlackPlayer(Player{
		ID:   uuid.New(),
		Name: bName,
	})
	a.write("\n")
	a.initGame()

	a.run()
}

func (a *Adapter) run() {
	for {
		a.write(a.Render())
		rawMove := a.requestForNextRawMove()
		err := a.processInput(rawMove)
		if err != nil {
			a.write(err.Error())
			continue
		}
		if a.g.State() != game.StateInProgress {
			break
		}
	}

	a.gameOver()
}

const inputDraw = "draw"
const inputUndo = "undo"

func (a *Adapter) processInput(input string) error {
	if input == inputDraw {
		err := a.g.ForceDraw()
		if err != nil {
			return err
		}
		return nil
	}

	if input == inputUndo {
		hasUndo := a.g.UndoLastMove()
		if hasUndo {
			a.write("undo successful")
		} else {
			a.write("no moves to undo")
		}
		return nil
	}

	err := a.g.ApplyMoveWithFileRank(input)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) writeIntro() {
	a.write("Welcome to CLI Chess")
}

func (a *Adapter) gameOver() {
	if a.g.State() == game.StateDraw || a.g.State() == game.StateStalemate {
		a.write(fmt.Sprintf("Game ended in a %v", a.g.State().String()))
		return
	}

	a.write(fmt.Sprintf("Winner: %v", a.player(a.g.Winner()).Name))
}

func (a *Adapter) requestForNextRawMove() string {
	return a.listenToNewInput(fmt.Sprintf("Player %s please enter input:", a.activePlayer().Name))
}

func (a *Adapter) requestForPlayerName(color engine.Color) string {
	return a.listenToNewInput(fmt.Sprintf("Please enter %s player's name:", color))
}

func (a *Adapter) setWhitePlayer(p Player) {
	a.whitePlayer = p
}

func (a *Adapter) setBlackPlayer(p Player) {
	a.blackPlayer = p
}

func (a *Adapter) activePlayer() Player {
	if a.g.ActiveColor() == engine.White {
		return a.whitePlayer
	}
	return a.blackPlayer
}

func (a *Adapter) player(c engine.Color) Player {
	if c == engine.White {
		return a.whitePlayer
	}
	return a.blackPlayer
}

func (a *Adapter) initGame() {
	a.g = game.NewGame(
		engine.NewBoard(),
	)
	a.write(fmt.Sprintf("Move input format:[fromFile][fromRank][toFile][toRank].\nie:a2a3\n\nWhite: %s | Black: %s\n", a.whitePlayer.Name, a.blackPlayer.Name))
}

func (a *Adapter) write(s string) {
	_, _ = a.writer.WriteString(s)
	_, _ = a.writer.WriteString("\n")
	_ = a.writer.Flush()
}

func (a *Adapter) listenToNewInput(prompt string) string {
	if prompt != "" {
		a.write(prompt)
	}

	input := ""
	if a.reader.Scan() {
		input = a.reader.Text()
	}

	if err := a.reader.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return input
}
