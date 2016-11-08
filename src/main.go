package main

import (
	"os"

	"github.com/BitFunnel/LabBook/src/cli"
)

func main() {
	cli.ParseAndDispatch(os.Args)
}
