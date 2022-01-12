package main

//
// SDL docs here
//   https://github.com/veandco/go-sdl2
// and
//   https://github.com/catsocks/sdl-grid/blob/master/main.c

import "github.com/veandco/go-sdl2/sdl"

const (
	displayX = 64
	displayY = 32

	// scaling factor
	cellSize = 16

	// scaled up window sizes
	windowX = int32(displayX * cellSize)
	windowY = int32(displayY * cellSize)
)

var (
	// colours
	background = sdl.Color{22, 22, 22, 255}
	lineColor  = sdl.Color{44, 44, 44, 255}
	pixelColor = sdl.Color{220, 220, 220, 220}
)

type Display struct {
	Pixels   [displayX][displayY]bool
	Rects    [displayX][displayY]sdl.Rect
	Renderer *sdl.Renderer
}

func SetupDisplay() *Display {
	d := Display{}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	//defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"Chip8",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		windowX,
		windowY,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}
	//defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	d.Renderer = renderer

	for x := 0; x < displayX; x++ {
		for y := 0; y < displayY; y++ {
			d.Rects[x][y] = sdl.Rect{
				int32(cellSize*x) + 1,
				int32(cellSize*y) + 1,
				cellSize - 2,
				cellSize - 2,
			}
		}
	}

	return &d
}

func (d *Display) Update() {
	// Draw background
	d.Renderer.SetDrawColor(
		background.R,
		background.G,
		background.B,
		background.A,
	)
	d.Renderer.Clear()

	// Draw lines
	d.Renderer.SetDrawColor(
		lineColor.R,
		lineColor.G,
		lineColor.B,
		lineColor.A,
	)
	for x := int32(0); x <= windowX; x += cellSize {
		d.Renderer.DrawLine(x, 0, x, windowY)
	}
	for y := int32(0); y <= windowY; y += cellSize {
		d.Renderer.DrawLine(0, y, windowX, y)
	}

	// Draw pixels
	d.Renderer.SetDrawColor(
		pixelColor.R,
		pixelColor.G,
		pixelColor.B,
		pixelColor.A,
	)
	for x, row := range d.Rects {
		for y := range row {
			if d.Pixels[x][y] {
				d.Renderer.FillRect(&d.Rects[x][y])
			}
		}
	}

	d.Renderer.Present()
}
