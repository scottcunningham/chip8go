package main

//
// Docs at:
// https://tobiasvl.github.io/blog/write-a-chip-8-emulator/#prerequisites
//

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	// cleanup
	"github.com/veandco/go-sdl2/sdl"
)

const (
	memoryBytes    = 4 * 1024
	numRegisters   = 16
	delayRate      = time.Second / 60 // 60Hz
	soundRate      = time.Second / 60 // 60Hz
	stackSize      = 16
	VF             = 0xF // flag register
	fontAddress    = 0x50
	programAddress = 0x200
	numKeys        = 16 // hex numpad
	KeyUnsupported = 0xff
)

var DefaultFont = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Chip8 struct {
	Memory [memoryBytes]uint8
	//.PixelsDisplay       [displayX][displayY]bool
	Display       *Display
	PC            uint16
	IndexRegister uint16
	Stack         [stackSize]uint16
	StackIndex    uint8
	Registers     [numRegisters]uint8
	DelayTimer    uint8
	SoundTimer    uint8
	timer         *time.Ticker
	rand          *rand.Rand
	Keypad        [numKeys]bool
}

func keyToKeyIndex(key sdl.Keycode) uint8 {
	switch key {
	case sdl.K_1:
		return 0x00
	case sdl.K_2:
		return 0x01
	case sdl.K_3:
		return 0x02
	case sdl.K_4:
		return 0x03
	case sdl.K_q:
		return 0x04
	case sdl.K_w:
		return 0x05
	case sdl.K_e:
		return 0x06
	case sdl.K_r:
		return 0x07
	case sdl.K_a:
		return 0x08
	case sdl.K_s:
		return 0x09
	case sdl.K_d:
		return 0x0a
	case sdl.K_f:
		return 0x0b
	case sdl.K_z:
		return 0x0c
	case sdl.K_x:
		return 0x0d
	case sdl.K_c:
		return 0x0e
	case sdl.K_v:
		return 0x0f
	default:
		return 0xff
	}
}

func (c *Chip8) PushButton(button sdl.Keycode) {
	idx := keyToKeyIndex(button)
	if idx == KeyUnsupported {
		return
	}
	c.Keypad[idx] = true
}

func (c *Chip8) ReleaseButton(button sdl.Keycode) {
	idx := keyToKeyIndex(button)
	if idx == KeyUnsupported {
		return
	}
	c.Keypad[idx] = false
}

func NewChip8() Chip8 {
	c := Chip8{}
	c.timer = time.NewTicker(delayRate)
	c.loadFont(DefaultFont)

	c.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	c.Display = SetupDisplay()

	c.PC = programAddress

	return c
}

func (c *Chip8) loadFont(font []uint8) {
	for i, x := range font {
		c.Memory[fontAddress+i] = x
	}
}

func repeatStr(s string, n int) string {
	ret := ""
	for x := 0; x < n; x++ {
		ret += s
	}
	return ret
}

func (c *Chip8) dumpMemory() {

	fmt.Println(repeatStr("=", 38) + " memory dump " + repeatStr("=", 38))
	for i, x := range c.Memory {
		if i%16 == 0 {
			fmt.Printf("\n 0x%03x |", i)
		}
		fmt.Printf(" 0x%02x", x)
	}
	fmt.Println()
	fmt.Println(repeatStr("=", 36) + " end memory dump " + repeatStr("=", 36))
}

func (c *Chip8) StackPush(val uint16) {
	if c.StackIndex == stackSize {
		panic("stack overflow")
	}
	c.StackIndex++
	c.Stack[c.StackIndex] = val
}

func (c *Chip8) StackPop() uint16 {
	val := c.Stack[c.StackIndex]
	c.StackIndex--
	return val
}

func (c *Chip8) Run(hz int) {
	done := make(chan bool)
	go c.RunTimers(done)
	go func() {
		for {
			c.Step()
			time.Sleep(time.Millisecond * time.Duration(1000/hz))
		}
		done <- true
	}()
	// can only do this from main thread, so other stuff is in goroutine
	c.ShowDisplay(done)
}

func (c *Chip8) ShowDisplay(done chan bool) {
	//
	// TODO: check chan

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				keyCode := t.Keysym.Sym
				if t.State == sdl.RELEASED {
					c.ReleaseButton(keyCode)
				} else if t.State == sdl.PRESSED {
					c.PushButton(keyCode)
				}
			}
		}
		c.Display.Update()
	}
}

func (c *Chip8) RunTimers(done chan bool) {
	for {
		select {
		case <-done:
			return
		case <-c.timer.C:
			if c.DelayTimer != 0 {
				c.DelayTimer -= 1
			}
			if c.SoundTimer != 0 {
				c.Beep()
				c.SoundTimer -= 1
			}
		}
	}
}

func (c *Chip8) Fetch() *Instruction {
	// fetch
	// - read instruction (x2) at mem addr pointed at by PC
	// - combine into 16 bit instr
	// - incr PC by 2
	rawInstr1 := c.Memory[c.PC]
	rawInstr2 := c.Memory[c.PC+1]
	rawInstr := uint16(rawInstr1)<<8 | uint16(rawInstr2)

	return &Instruction{
		Prefix: uint8(0b00001111 & (rawInstr >> 12)),
		X:      uint8(0b00001111 & (rawInstr >> 8)),
		Y:      uint8(0b00001111 & (rawInstr >> 4)),
		N:      uint8((0b00001111 & rawInstr)),
		NN:     uint8((0b11111111 & rawInstr)),
		NNN:    uint16(0b0000111111111111 & uint16(rawInstr)),
		Raw:    rawInstr,
	}
}

// consider using https://pkg.go.dev/github.com/thinkofdeath/steven/type/nibble
type Instruction struct {
	Prefix      uint8
	X, Y, N, NN uint8
	NNN         uint16
	Raw         uint16
}

func (c *Chip8) Step() {
	instr := c.Fetch()

	// decode & execute
	// stuff is split into nibbles (4 bits)
	// extract the following:
	// - prefix (nib1)
	// - X (nib2) register index
	// - Y (nib3) register index
	// - N (nib4) just a 4-bit number
	// - NN (second byte) immed number
	// - NNN (nib2+3+4) memory addr
	fmt.Printf("instr 0x%0X pc=0x%0X %+v\n", instr.Raw, c.PC, instr)
	switch instr.Prefix {
	case 0x0:
		// 00E0 - clear screen
		if instr.Y != 0xE {
			panic(fmt.Sprintf("bad instr %x at addr 0x%x", instr.Raw, c.PC))
		}
		c.Clear()
	case 0x1:
		// 1NNN - jump - set PC to NNN
		c.PC = instr.NNN
		// don't increment PC
		return
	case 0x2:
		// 2NNN - call - set PC to NNN, but push current PC to stack first
		c.StackPush(c.PC)
		c.PC = instr.NNN
		// don't increment PC
		return
	case 0x3:
		// 3XNN - skip one instruction if rX == NN
		if c.Registers[instr.X] == instr.NN {
			c.PC += 2
		}
	case 0x4:
		// 4XNN - skip one instruction if rX != NN
		if c.Registers[instr.X] != instr.NN {
			c.PC += 2
		}
	case 0x5:
		// 5XY0 - skip one instruction if rX == rY
		if c.Registers[instr.X] == c.Registers[instr.Y] {
			c.PC += 2
		}
	case 0x6:
		// 6XNN - set rX to NN
		c.Registers[instr.X] = instr.NN
	case 0x7:
		// 7XNN - add NN to rX -- DOES NOT AFFECT CARRY FLAG
		c.Registers[instr.X] += instr.NN
	case 0x8:
		switch instr.N {
		// set
		case 0:
			//   8XY0 - set rX to value of rY
			c.Registers[instr.X] = c.Registers[instr.Y]
		// bitwise
		case 1:
			//   8XY1 - set rX to value of (rX | rY)
			c.Registers[instr.X] = c.Registers[instr.X] | c.Registers[instr.Y]
		case 2:
			//   8XY2 - set rX to value of (rX & rY)
			c.Registers[instr.X] = c.Registers[instr.X] & c.Registers[instr.Y]
		case 3:
			//   8XY3 - set rX to value of (rX ^ rY)
			c.Registers[instr.X] = c.Registers[instr.X] ^ c.Registers[instr.Y]
		// add
		case 4:
			//   8XY4 - set rX to value of (rX + rY) - AFFECTS CARRY FLAG unlike 7XNN
			x, y := c.Registers[instr.X], c.Registers[instr.Y]
			result := x + y
			c.Registers[instr.X] = result
			if result < x {
				// we overflowed
				c.Registers[VF] = 1
			} else {
				c.Registers[VF] = 0
			}
		// subtract
		case 5:
			//   8XY5 - set rX to value of rX - rY - AFFECTS CARRY FLAG IN AMBIGUOUS WAY
			x, y := c.Registers[instr.X], c.Registers[instr.Y]
			c.Registers[instr.X] = x - y
			if x > y {
				c.Registers[VF] = 1
			} else {
				c.Registers[VF] = 0
			}
		case 7:
			//   8XY7 - set rX to vlaue of rY - rX - AFFECTS CARRY FLAG IN AMBIGUOUS WAY
			x, y := c.Registers[instr.X], c.Registers[instr.Y]
			c.Registers[instr.X] = y - x
			if y > x {
				c.Registers[VF] = 1
			} else {
				c.Registers[VF] = 0
			}
		// shift
		case 6:
			switch instr.N {
			//   8XY6 - shift right
			case 0x6:
				// TODO: implement original version where rX is set to rY before continuing
				x := c.Registers[instr.X]
				overflow := (x & 0b00000001)
				c.Registers[instr.X] = x >> 1
				c.Registers[VF] = overflow

			//   8XYE - shift left
			case 0xE:
				// TODO: implement original version where rX is set to rY before continuing
				x := c.Registers[instr.X]
				overflow := (x & 0b10000000) >> 7
				c.Registers[instr.X] = x << 1
				c.Registers[VF] = overflow

			default:
				panic(fmt.Sprintf("bad shift instr %X at addr 0x%x", instr.Raw, c.PC))
			}
		}
	case 0x9:
		// 9XY0 - skip one instruction if rX != rY
		if c.Registers[instr.X] != c.Registers[instr.Y] {
			c.PC += 2
		}
	case 0xA:
		// ANNN - set Index Register to NNN
		c.IndexRegister = instr.NNN
	case 0xB:
	case 0xC:
		// CXNN - generate random number, AND it with NN, put it into rX
		val := uint8(c.rand.Intn(0xFF))
		c.Registers[instr.X] = val & instr.NN
	case 0xD:
		// DXYN - draw
		x := c.Registers[instr.X] % displayX
		y := c.Registers[instr.Y] % displayY
		n := instr.N
		row := uint8(0)

		c.Registers[VF] = 0
		for yLine := uint8(0); yLine < n; yLine++ {
			spriteAddr := c.IndexRegister + uint16(yLine)
			row = c.Memory[spriteAddr]
			for xLine := uint8(0); xLine < 8; xLine++ {
				currentPixelIsOn := (row&(0b10000000>>xLine) != 0)
				if currentPixelIsOn {
					// If the current pixel in the sprite row is on and the pixel at
					// coordinates X,Y on the screen is also on, turn off the pixel and set VF to 1
					if c.Display.Pixels[x+xLine][y+yLine] {
						c.Display.Pixels[x+xLine][y+yLine] = false
						c.Registers[VF] = 1
					} else {
						// Or if the current pixel in the sprite row is on and the screen pixel is not,
						// draw the pixel at the X and Y coordinates
						c.Display.Pixels[x+xLine][y+yLine] = true
					}
					// TODO: check for edge of screen!
				}
			}
		}
	case 0xE:
		// TODO: user input stuff
		switch instr.NN {
		case 0x9E:
			// EX9E -- skip instruction if keys[val of rX] is pressed
			keyIdx := c.Registers[instr.X]
			if c.Keypad[keyIdx] {
				c.PC += 2
			}
			break
		// EXA1
		case 0xA1:
			// EXA1 -- skip instruction if keys[val of rX] is NOT pressed
			keyIdx := c.Registers[instr.X]
			if !c.Keypad[keyIdx] {
				c.PC += 2
			}
			break
		default:
			panic(fmt.Sprintf("bad input instr %X at addr 0x%x", instr.Raw, c.PC))
		}
	case 0xF:
		// Timers
		switch instr.NN {
		// FX07 - set rX to current value of delay timer
		case 0x07:
			c.Registers[instr.X] = c.DelayTimer
		// FX15 - set delay timer to current value of rX
		case 0x15:
			c.DelayTimer = c.Registers[instr.X]
		// FX18 - set sound timer to current value of rX
		case 0x18:
			c.SoundTimer = c.Registers[instr.X]
		default:
			c.dumpMemory()
			panic(fmt.Sprintf("bad timer instr %X at addr 0x%x", instr.Raw, c.PC))
		}
	default:
		panic(fmt.Sprintf("bad instr %X at addr 0x%x", instr.Raw, c.PC))
	}

	// PC increments by 2 since instructions are 16-bit, not 8
	c.PC += 2
}

func (c *Chip8) Clear() {
	for x, arr := range c.Display.Pixels {
		for y := range arr {
			c.Display.Pixels[x][y] = false
		}
	}
}

func (c *Chip8) Beep() {
	// TODO beep
	fmt.Print("\a")
}

func (c *Chip8) Load(program []uint8) {
	for i, v := range program {
		c.Memory[programAddress+i] = v
	}
	c.PC = programAddress
}

func (c *Chip8) LoadFromFile(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	c.Load(bytes)
	return nil
}
