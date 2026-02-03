package main

import "github.com/dyxj/chess/internal/adapter/cli"

func main() {
	adapter := cli.NewAdapter()
	adapter.Run()
}
