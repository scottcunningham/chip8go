package main

import (
	"github.com/jessevdk/go-flags"
)

type Args struct {
	ROM string `short:"r" long:"rom" description:"Path to chip-8 ROM to execute" required:"yes"`
}

func main() {
	var args Args
	_, err := flags.Parse(&args)
	if err != nil {
		panic(err)
	}

	chip := NewChip8()
	chip.LoadFromFile(args.ROM)
	chip.Run(6)
}
