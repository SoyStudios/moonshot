package main

import (
	"fmt"
	"sync"
)

type (
	stack []int16

	// Machine is the stack machine powering bots.
	//
	// It is a 16 bit stack machine with 16 persistent
	// registers.
	//
	// state represents the interface to the bot.
	Machine struct {
		run runFunc

		pc int // program counter

		program   []*Gene
		stack     *stack
		registers [16]int16

		state State
	}

	runFunc func(*Machine, []int16, func())

	// Gene represents one gene of the bot's program.
	//
	// A Gene consists of two sections, an evaluation
	// and an execution section.
	// By the end of the evaluation section, the stack
	// will be popped. If the value is > 0 the execution
	// section will be executed.
	Gene struct {
		Evaluate []int16
		Execute  []int16
	}

	State interface {
		Reset()
		X() int16
		Y() int16
		Energy() int16
		ID() int16
		RemoteID(int16) int16
		Scan(int16, int16) (int16, int16)
		Thrust(int16, int16)
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
	m.state.Reset()
	for _, g := range m.program {
		m.RunGene(g)
	}
}

func (m *Machine) RunGene(g *Gene) {
	m.pc = 0
	m.stack.Reset()
	if len(g.Evaluate) == 0 {
		return
	}
	for {
		inst := Translate(Token(g.Evaluate[m.pc]))
		inst.Run(m, g.Evaluate)
		m.pc++
		if m.pc > len(g.Evaluate)-1 {
			break
		}
	}
	if len(*m.stack) <= 0 {
		return
	}
	if m.stack.Pop() < 1 {
		return
	}

	if len(g.Execute) == 0 {
		return
	}
	m.pc = 0
	m.stack.Reset()
	for {
		inst := Translate(Token(g.Execute[m.pc]))
		inst.Run(m, g.Execute)
		m.pc++
		if m.pc > len(g.Execute)-1 {
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
	var ran bool
	return func(m *Machine, code []int16, f func()) {
		if !ran && m.pc == breakpoint {
			ran = true
			if !breakFunc(m) {
				return
			}
		}
		runFunc(m, code, f)
	}
}
