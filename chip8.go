package main

//
// Docs at:
// https://tobiasvl.github.io/blog/write-a-chip-8-emulator/#prerequisites
//

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	memoryBytes  = 4 * 1024
	displayX     = 64
	displayY     = 32
	numRegisters = 16
	delayRate    = time.Second / 60 // 60Hz
	soundRate    = time.Second / 60 // 60Hz
	stackSize    = 16
	VF           = 0xF // flag register
)

var DefaultFont = [][]uint8{
	[]uint8{0xF0, 0x90, 0x90, 0x90, 0xF0}, // 0
	[]uint8{0x20, 0x60, 0x20, 0x20, 0x70}, // 1
	[]uint8{0xF0, 0x10, 0xF0, 0x80, 0xF0}, // 2
	[]uint8{0xF0, 0x10, 0xF0, 0x10, 0xF0}, // 3
	[]uint8{0x90, 0x90, 0xF0, 0x10, 0x10}, // 4
	[]uint8{0xF0, 0x80, 0xF0, 0x10, 0xF0}, // 5
	[]uint8{0xF0, 0x80, 0xF0, 0x90, 0xF0}, // 6
	[]uint8{0xF0, 0x10, 0x20, 0x40, 0x40}, // 7
	[]uint8{0xF0, 0x90, 0xF0, 0x90, 0xF0}, // 8
	[]uint8{0xF0, 0x90, 0xF0, 0x10, 0xF0}, // 9
	[]uint8{0xF0, 0x90, 0xF0, 0x90, 0x90}, // A
	[]uint8{0xE0, 0x90, 0xE0, 0x90, 0xE0}, // B
	[]uint8{0xF0, 0x80, 0x80, 0x80, 0xF0}, // C
	[]uint8{0xE0, 0x90, 0x90, 0x90, 0xE0}, // D
	[]uint8{0xF0, 0x80, 0xF0, 0x80, 0xF0}, // E
	[]uint8{0xF0, 0x80, 0xF0, 0x80, 0x80}, // F
}

type Chip8 struct {
	Memory        [memoryBytes]uint8
	Display       [displayX][displayY]bool
	PC            uint16
	IndexRegister uint16
	Stack         [stackSize]uint16
	StackIndex    uint8
	Registers     [numRegisters]uint8
	DelayTimer    uint8
	SoundTimer    uint8
	timer         *time.Ticker
	Font          [][]uint8
	rand          *rand.Rand
}

func NewChip8() Chip8 {
	c := Chip8{}
	c.timer = time.NewTicker(delayRate)
	c.Font = DefaultFont

	c.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	return c
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

func (c *Chip8) Run() {
	done := make(chan bool)
	go c.RunTimers(done)
	go c.ShowDisplay(done)

	for {
		// FIXME ugly lol
		if c.PC == 4096 {
			c.PC = 0
		}
		c.Step()
	}

	done <- true
	return
}

func (c *Chip8) ShowDisplay(done chan bool) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// FIXME(nasty)
			for x, arr := range c.Display {
				for y := range arr {
					val := " "
					if c.Display[x][y] {
						val = "x"
					}
					fmt.Printf("%s", val)
				}
				fmt.Printf("\n")
			}
			fmt.Printf("===========================\n")
		}
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
	fmt.Printf("%x %x %x\n", rawInstr1, rawInstr2, rawInstr)

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
	fmt.Printf("%+v %x\n", instr, instr.Raw)

	// decode & execute
	// stuff is split into nibbles (4 bits)
	// extract the following:
	// - prefix (nib1)
	// - X (nib2) register index
	// - Y (nib3) register index
	// - N (nib4) just a 4-bit number
	// - NN (second byte) immed number
	// - NNN (nib2+3+4) memory addr
	switch instr.Prefix {
	case 0x0:
		// 00E0 - clear screen
		if instr.Y != 0xE {
			panic("bad instr" + fmt.Sprintf("%x", instr.Raw))
		}
		c.Clear()
		break
	case 0x1:
		// 1NNN - jump - set PC to NNN
		c.PC = instr.NNN
		break
	case 0x2:
		// 2NNN - call - set PC to NNN, but push current PC to stack first
		c.StackPush(c.PC)
		c.PC = instr.NNN
		break
	case 0x3:
		// 3XNN - skip one instruction if rX == NN
		if c.Registers[instr.X] == instr.NN {
			c.PC += 2
		}
		break
	case 0x4:
		// 4XNN - skip one instruction if rX != NN
		if c.Registers[instr.X] != instr.NN {
			c.PC += 2
		}
		break
	case 0x5:
		// 5XY0 - skip one instruction if rX == rY
		if c.Registers[instr.X] == c.Registers[instr.Y] {
			c.PC += 2
		}
		break
	case 0x6:
		// 6XNN - set rX to NN
		c.Registers[instr.X] = instr.NN
		break
	case 0x7:
		// 7XNN - add NN to rX -- DOES NOT AFFECT CARRY FLAG
		c.Registers[instr.X] += instr.NN
		break
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
				panic("bad shift")
			}
		}
		break
	case 0x9:
		// 9XY0 - skip one instruction if rX != rY
		if c.Registers[instr.X] != c.Registers[instr.Y] {
			c.PC += 2
		}
		break
	case 0xA:
		// ANNN - set Index Register to NNN
		c.IndexRegister = instr.NNN
		break
	case 0xB:
		break
	case 0xC:
		// CXNN - generate random number, AND it with NN, put it into rX
		val := uint8(c.rand.Intn(0xFF))
		c.Registers[instr.X] = val & instr.NN
		break
	case 0xD:
		// DXYN - draw
		// FIXME: mvp
		x := c.Registers[instr.X] % displayX
		y := c.Registers[instr.Y] % displayY
		n := instr.N

		c.Registers[VF] = 0
		for i := uint8(0); i < n; i++ {
			spriteAddr := c.IndexRegister + uint16(i)
			fmt.Printf("rm me %d%d%d%d\n", x, y, n, spriteAddr)
		}
		break
	case 0xE:
		// TODO: user input stuff
		break
	case 0xF:
		// Timers
		switch instr.NN {
		// FX07 - set rX to current value of delay timer
		case 0x07:
			c.Registers[instr.X] = c.DelayTimer
			break
		// FX15 - set delay timer to current value of rX
		case 0x15:
			c.DelayTimer = c.Registers[instr.X]
			break
		// FX18 - set sound timer to current value of rX
		case 0x18:
			c.SoundTimer = c.Registers[instr.X]
			break
		default:
			panic("bad timer instr")
		}
		break
	default:
		panic("bad instruction")
	}

	// PC increments by 2 since instructions are 16-bit, not 8
	c.PC += 2
}

func (c *Chip8) Clear() {
	for x, arr := range c.Display {
		for y := range arr {
			c.Display[x][y] = false
		}
	}
}

func (c *Chip8) Beep() {
	// TODO beep
	fmt.Print("\a")
}

func (c *Chip8) Load(program []uint8) {
	for i, v := range program {
		c.Memory[i] = v
	}
}

func main() {
	fmt.Println("vim-go")
	chip := NewChip8()
	chip.Load(IBMLogo)
	chip.Run()
}
