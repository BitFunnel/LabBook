package main

import (
	"os"

	"github.com/bitfunnel/LabBook/src/cli"
)

func main() {
	cli.ParseAndDispatch(os.Args)
}
