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

	stateMock.On("Energy").Return(int16(1))
	stateMock.On("Y").Return(int16(2))
	stateMock.On("Scan", int16(1), int16(2)).Return(int16(16), int16(17))

	m.Run()
	t.Logf("%v", program)

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
END
BEGIN EX
	RID
	RDX
	RDY
	SCN
	ABS
	PSH CON 1024
	LEQ
	PSH CON 0
	GEQ
	PSH CON 1
	XOR
	PSH CON 2
	SUB
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
	m.run = runWithBreak(44, runInstructionDebug)
	m.program = program
	stateMock := &StateMock{}
	m.state = stateMock

	stateMock.On("X").Return(int16(42))
	stateMock.On("Y").Return(int16(420))
	stateMock.On("Energy").Return(int16(17))

	m.Run()
	t.Logf("%v", program)
	if !assert.Equal(t, 46, m.pc) {
		return
	}

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

	stateMock.On("Energy").Return(int16(1000))
	stateMock.On("Reproduce", int16(500)).Once()

	m.Run()
	t.Logf("%v", program)

	if !stateMock.AssertExpectations(t) {
		return
	}
}
