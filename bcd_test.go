package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ir = 0x500

func TestBCDOpcode(t *testing.T) {
	c := NewChip8()
	// FX33 - binary-coded decimal conversion

	cases := []struct {
		x      uint8
		output [3]uint8
	}{
		{
			x:      123,
			output: [3]uint8{1, 2, 3},
		},
		{
			x:      10,
			output: [3]uint8{0, 1, 0},
		},
		{
			x:      255,
			output: [3]uint8{2, 5, 5},
		},
		{
			x:      0,
			output: [3]uint8{0, 0, 0},
		},
		{
			x:      7,
			output: [3]uint8{0, 0, 7},
		},
	}

	for idx, input := range cases {
		c.PC = programAddress             // reset PC
		c.Memory[programAddress] = 0xF1   // FX
		c.Memory[programAddress+1] = 0x33 // 33

		c.IndexRegister = ir
		c.Registers[1] = input.x
		c.Step()

		assert.Equal(t, c.Memory[ir], input.output[0], fmt.Sprintf("case #%d", idx))
		assert.Equal(t, c.Memory[ir+1], input.output[1], fmt.Sprintf("case #%d", idx))
		assert.Equal(t, c.Memory[ir+2], input.output[2], fmt.Sprintf("case #%d", idx))
	}
}
