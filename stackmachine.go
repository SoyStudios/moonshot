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

	runFunc func(*Machine, []int16, func())

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
	for {
		inst := Translate(Token(m.program.Evaluate[m.pc]))
		inst.Run(m, m.program.Evaluate)
		m.pc++
		if m.pc > len(m.program.Evaluate)-1 {
			break
		}
	}
	if len(*m.stack) <= 0 {
		return
	}
	if m.stack.Pop() < 1 {
		return
	}

	if len(m.program.Execute) == 0 {
		return
	}
	m.pc = 0
	m.stack.Reset()
	for {
		inst := Translate(Token(m.program.Execute[m.pc]))
		inst.Run(m, m.program.Execute)
		m.pc++
		if m.pc > len(m.program.Execute)-1 {
			break
		}
	}
}

func runInstruction(m *Machine, code []int16, f func()) {
	f()
}

func runInstructionDebug(m *Machine, code []int16, f func()) {
	fmt.Printf("%v\n", code)
	fmt.Println("pc", m.pc)
	inst := Translate(Token(code[m.pc]))
	fmt.Printf("%s\n", inst.String(m, code))
	fmt.Println("exec")

	f()

	fmt.Printf("stack: %v\n\n", m.stack)
}

func runWithBreak(breakpoint int, breakFunc func(m *Machine) bool, runFunc runFunc) runFunc {
	return func(m *Machine, code []int16, f func()) {
		if m.pc == breakpoint {
			if !breakFunc(m) {
				return
			}
		}
		runFunc(m, code, f)
	}
}
