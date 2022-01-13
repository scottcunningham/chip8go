package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShiftRightOpcode(t *testing.T) {
	c := NewChip8()
	// 8XY6 - shift right

	cases := []struct {
		x     uint8
		carry uint8
		out   uint8
	}{
		{
			x:     4,
			carry: 0,
			out:   4 >> 1,
		},
		{
			x:     1,
			carry: 1,
			out:   1 >> 1,
		},
	}

	for idx, input := range cases {
		c.PC = programAddress             // reset PC
		c.Memory[programAddress] = 0x81   // 8X
		c.Memory[programAddress+1] = 0x26 // Y6

		c.Registers[1] = input.x
		c.Step()

		assert.Equal(t, c.Registers[1], input.out, fmt.Sprintf("case #%d", idx))
	}
}

func TestShiftLeftOpcode(t *testing.T) {
	c := NewChip8()
	// 8XYE - shift right

	cases := []struct {
		x     uint8
		carry uint8
		out   uint8
	}{
		{
			x:     0b00000100,
			carry: 0,
			out:   0b00001000,
		},
		{
			x:     0b11111111,
			carry: 1,
			out:   0b11111110,
		},
	}

	for idx, input := range cases {
		c.PC = programAddress             // reset PC
		c.Memory[programAddress] = 0x81   // 8X
		c.Memory[programAddress+1] = 0x2E // YE

		c.Registers[1] = input.x
		c.Step()

		assert.Equal(t, c.Registers[1], input.out, fmt.Sprintf("case #%d", idx))
	}
}
