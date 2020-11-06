package main

import "testing"

func TestSimpleMachine(t *testing.T) {
	program := []Token{
		BGN,
		RDX,
		PSH, CON, 0,
		GEQ,
		EXE,
		PSH, CON, 12,
		THR,
		END,
	}
	m := NewMachine()
	m.program = Program(program)
}
