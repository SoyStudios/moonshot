package main

import "sync"

type (
	stack []int16

	Machine struct {
		pc int // program counter
		i  int // instruction counter

		program   Program
		stack     stack
		callStack stack
		registers [16]int16

		state State
	}

	Program struct {
		Evaluate []int16
		Execute  []int16
	}

	State interface {
		X() int16
		Y() int16
		Angle() int16
		Energy() int16
		ID() int16
		RemoteID(int16) int16
		Scan(int16) int16
		Thrust(int16)
		Turn(int16)
		Mine(int16)
		Reproduce(int16)
	}
)

var (
	stackPool = sync.Pool{
		New: func() interface{} {
			return stack(make([]int16, 0, 16))
		},
	}
)

func (s stack) Push(v int16) {
	s = append(s, v)
}

func (s stack) Pop() int16 {
	n := len(s) - 1
	ret := s[n]
	s = s[:n]
	return ret
}

func NewMachine() *Machine {
	m := &Machine{
		callStack: stackPool.Get().(stack)[:0],
		stack:     stackPool.Get().(stack)[:0],
	}
	return m
}

func (m *Machine) Run() {
	m.pc = 0
	if len(m.program.Evaluate) == 0 {
		return
	}
	inst := Translate(Token(m.program.Evaluate[m.pc]))
	inst.Run(m, m.program.Evaluate)
}
