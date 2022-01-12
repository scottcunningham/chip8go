package main

func main() {
	chip := NewChip8()
	chip.LoadFromFile("IBMLogo.c8")
	chip.Run(6)
}
