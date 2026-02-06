package main

import (
	"flag"
	"fmt"

	"github.com/dyxj/chess/internal/adapter/cli"
)

const (
	iconSymbol = "symbol"
	iconNumber = "number"
)

func main() {
	icon := flag.String("icon", "symbol", "determines icons used to represent pieces")

	flag.Parse()

	var opts []cli.Option

	switch *icon {
	case iconSymbol:
		opts = append(opts, cli.WithSymbolIconMapper())
	case iconNumber:
		opts = append(opts, cli.WithNumberIconMapper())
	default:
		fmt.Println("invalid icon flag")
		return
	}

	adapter := cli.NewAdapter(opts...)

	adapter.Run()
}
