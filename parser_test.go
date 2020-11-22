package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleParser(t *testing.T) {
	code := `
BEGIN EV
	// if energy >= 5 and y >= 5
	PSH CON 5
	RDE
	GEQ
	PSH CON 5
	RDY
	GEQ
	AND
END
BEGIN EX
	// push 1, 2
	PSH CON 1
	PSH CON 2
	SCN
	POP REG 1
	POP REG 0
	PSH REG 0
	PSH REG 1
	AND
END

BEGIN EV
	PSH CON 1
END
BEGIN EX
	PSH CON 120
END
	`
	p := NewParser(strings.NewReader(code))
	program, err := p.Parse()
	if err != nil {
		t.Fatal(err)
		return
	}

	m := NewMachine()
	m.run = runInstruction
	m.program = program
	stateMock := &StateMock{}
	m.state = stateMock

	stateMock.On("Energy").Return(int16(1))
	stateMock.On("Y").Return(int16(2))
	stateMock.On("Scan", int16(1), int16(2)).Return(int16(16), int16(17))

	m.RunGene(program[0])

	if !assert.Len(t, *m.stack, 1, "stack size") {
		return
	}
	if !assert.Equal(t, int16(16&17), (*m.stack)[0]) {
		return
	}
	if !assert.Equal(t, int16(16), m.registers[0]) {
		return
	}
	if !stateMock.AssertExpectations(t) {
		return
	}

	m.RunGene(program[1])
	if !assert.Equal(t, int16(120), (*m.stack)[0]) {
		return
	}
}

func TestFullLanguage(t *testing.T) {
	code := `
BEGIN EV
	RDX
	RDY
	ABS
	RDE
	PSH CON 0
	GRT
	PSH CON 2
	POP REG 0
	PSH CON 3
	POP REG 1
	PSH REG 0
	PSH REG 1
	LST
	AND
	XOR
	NOT
	PSH CON 0
	IOR
	PSH REG 1024
	PSH CON 2
	PSH REG 0
	IEQ
	NOP
END
BEGIN EX
	RID
	RDX
	RDY
	SCN
	ABS
	PSH CON -23
	LEQ
	PSH CON 0
	GEQ
	PSH CON 1
	XOR
	PSH CON 2
	SUB
	PSH CON 134
	THR
	RDX
	NEG
	RDY
	NEG
	TRN
	PSH REG 1
	MNE
	PSH REG 0
	PSH REG 1
	ADD
	PSH CON 1
	ADD
	PSH CON 2
	MUL
	PSH CON 3
	DIV
	REP
END
	`
	p := NewParser(strings.NewReader(code))
	program, err := p.Parse()
	if err != nil {
		t.Fatal(err)
		return
	}

	m := NewMachine()
	m.run = runWithBreak(44, func(m *Machine) bool {
		if !assert.Equal(t, 44, m.pc) {
			return false
		}
		if !assert.Equal(t, int16(0), (*m.stack)[0]) {
			return false
		}
		if !assert.Equal(t, int16(1), (*m.stack)[1]) {
			return false
		}
		if !assert.Equal(t, int16(2), m.registers[0]) {
			return false
		}
		if !assert.Equal(t, int16(3), m.registers[1]) {
			return false
		}
		return true
	},
		runInstructionDebug)
	m.program = program
	stateMock := &StateMock{}
	m.state = stateMock

	stateMock.On("Reset")
	stateMock.On("Execute")
	stateMock.On("X").Return(int16(42))
	stateMock.On("Y").Return(int16(420))
	stateMock.On("Energy").Return(int16(17))
	stateMock.On("Scan", int16(42), int16(420)).Return(int16(12), int16(34))
	stateMock.On("Thrust", int16(-2), int16(134))
	stateMock.On("Turn", int16(-420))
	stateMock.On("Mine", int16(3))
	stateMock.On("Reproduce", int16(4))

	m.Run()

	if !stateMock.AssertExpectations(t) {
		return
	}
}

func TestTutorialBot(t *testing.T) {
	code := `
BEGIN EV
	// Read bot’s current energy level and push it to the stack
	RDE
	PSH CON 1000
	GEQ
END
BEGIN EX
	PSH CON 500
	REP
END

BEGIN EV
	// If total velocity is >= 400
	RDX
	RDY
	ABS
	PSH CON 400
	GEQ
END
BEGIN EX
	// Turn and thrust in opposite direction
	RDX
	NEG
	RDY
	NEG
	THR
END
	`
	p := NewParser(strings.NewReader(code))
	program, err := p.Parse()
	if err != nil {
		t.Fatal(err)
		return
	}

	m := NewMachine()
	m.run = runInstructionDebug
	m.program = program
	stateMock := &StateMock{}
	m.state = stateMock

	stateMock.On("Reset")
	stateMock.On("Execute")
	stateMock.On("Energy").Return(int16(1000))
	stateMock.On("Reproduce", int16(500)).Once()
	stateMock.On("X").Return(int16(42))
	stateMock.On("Y").Return(int16(420))
	stateMock.On("Thrust", int16(-42), int16(-420))

	m.Run()
	t.Logf("%v", program)

	if !stateMock.AssertExpectations(t) {
		return
	}
}
