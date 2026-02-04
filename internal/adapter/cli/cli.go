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
}

func NewAdapter() *Adapter {
	return &Adapter{
		reader: bufio.NewScanner(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
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
	a.initGame()

	a.run()
}

func (a *Adapter) run() {
	for {
		a.write(a.Render())
		rawMove := a.requestForNextRawMove()
		isCheckMate, err := a.processInput(rawMove)
		if err != nil {
			a.write(err.Error())
		}
		if isCheckMate {
			break
		}
	}

	a.gameOver()
}

func (a *Adapter) writeIntro() {
	a.write("Welcome to CLI Chess")
}

func (a *Adapter) gameOver() {
	a.write("Game Over")
}

func (a *Adapter) processInput(input string) (isCheckMate bool, err error) {

	a.write(fmt.Sprintf("processing %v---\n", input))

	if input == "x" {
		isCheckMate = true
	}

	return isCheckMate, nil
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

func (a *Adapter) initGame() {
	a.g = game.NewGame(
		engine.NewBoard(),
	)
	a.write(fmt.Sprintf("\nWhite: %s | Black: %s\n", a.whitePlayer.Name, a.blackPlayer.Name))
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
