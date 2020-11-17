package main

import (
	"fmt"
	"sync"
)

type (
	stack []int16

	Machine struct {
		run func(*Machine, []int16, func())

		pc int // program counter

		program   *Program
		stack     *stack
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
		Energy() int16
		ID() int16
		RemoteID(int16) int16
		Scan(int16, int16) (int16, int16)
		Thrust(int16)
		Turn(int16, int16)
		Mine(int16)
		Reproduce(int16)
	}
)

var (
	stackPool = sync.Pool{
		New: func() interface{} {
			s := stack(make([]int16, 0, 16))
			return &s
		},
	}
)

func (s *stack) Reset() {
	*s = (*s)[:0]
}

func (s *stack) Push(v int16) {
	*s = append(*s, v)
}

func (s *stack) Pop() int16 {
	n := len(*s) - 1
	ret := (*s)[n]
	*s = (*s)[:n]
	return ret
}

func NewMachine() *Machine {
	m := &Machine{
		run:   runInstruction,
		stack: stackPool.Get().(*stack),
	}
	return m
}

func (m *Machine) Destroy() {
	stackPool.Put(m.stack)
	m.stack = nil
}

func (m *Machine) Run() {
	m.pc = 0
	m.stack.Reset()
	if len(m.program.Evaluate) == 0 {
		return
	}
	inst := Translate(Token(m.program.Evaluate[m.pc]))
	inst.Run(m, m.program.Evaluate)
	if len(*m.stack) <= 0 {
		return
	}
	if m.stack.Pop() < 1 {
		return
	}
	m.pc = 0
	m.stack.Reset()
	inst = Translate(Token(m.program.Execute[m.pc]))
	inst.Run(m, m.program.Execute)
}

func runInstruction(m *Machine, code []int16, f func()) {
	m.pc++
	f()
	if m.pc > len(code)-1 {
		return
	}
	inst := Translate(Token(code[m.pc]))
	inst.Run(m, code)
}

func runInstructionDebug(m *Machine, code []int16, f func()) {
	fmt.Printf("%v\n", code)
	fmt.Println("pc", m.pc)

	m.pc++
	f()
	fmt.Printf("%v\n\n", m.stack)
	if m.pc > len(code)-1 {
		return
	}
	tok := Token(code[m.pc])
	fmt.Printf("%s\n", tok)

	inst := Translate(tok)
	inst.Run(m, code)
}

func runWithBreak(breakpoint int, runFunc func(*Machine, []int16, func())) func(*Machine, []int16, func()) {
	return func(m *Machine, code []int16, f func()) {
		m.pc++
		if m.pc == breakpoint {
			m.pc--
			return
		}
		m.pc--
		runFunc(m, code, f)
	}
}
