package main

func main() {
	chip := NewChip8()
	chip.Load(IBMLogo)

	chip.Display.Pixels[3][25] = true
	chip.Display.Pixels[15][30] = true
	chip.Display.Pixels[4][12] = true
	//chip.ShowDisplay(make(chan bool))
	chip.Run(6)
}
