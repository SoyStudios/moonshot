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

	if !assert.Len(t, m.stack, 1, "stack size") {
		return
	}
	if !assert.Equal(t, int16(16&17), m.stack[0]) {
		return
	}
	if !assert.Equal(t, int16(16), m.registers[0]) {
		return
	}
}
