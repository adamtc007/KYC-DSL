package main

import (
	"os"

	"github.com/adamtc007/KYC-DSL/internal/cli"
)

func main() {
	cli.Run(os.Args[1:])
}
