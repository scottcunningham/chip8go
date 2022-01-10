package main

import "github.com/veandco/go-sdl2/sdl"

const (
	cellSize = 24
)

func display() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	windowX := displayX * cellSize
	windowY := displayY * cellSize

	window, err := sdl.CreateWindow(
		"Chip8",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		windowX,
		windowY,
		sdl.WINDOW_SHOWN
	)
	if err != nil {
		panic(err)
	}
}
