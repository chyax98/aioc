package main

import (
	"os"

	"aioc/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(os.Args[1:], version); err != nil {
		cli.PrintError(err)
		os.Exit(1)
	}
}
