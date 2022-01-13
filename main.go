package main

import (
	"github.com/jessevdk/go-flags"
	"os"
	"runtime/pprof"
)

type Args struct {
	ROM         string  `short:"r" long:"rom" description:"Path to chip-8 ROM to execute" required:"yes"`
	RefreshRate int     `long:"refresh-rate" description:"Cycles per second" default:"60"`
	DebugMode   bool    `short:"d" long:"debug" description:"Print verbose data about instructions"`
	Profile     *string `long:"profile" description:"Dump a CPU profile into the given file"`
}

func main() {
	var args Args
	_, err := flags.Parse(&args)
	if err != nil {
		panic(err)
	}

	if args.Profile != nil {
		f, err := os.Create(args.Profile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	chip := NewChip8()
	chip.DebugMode = args.DebugMode
	chip.LoadFromFile(args.ROM)
	chip.Run(args.RefreshRate)
}
