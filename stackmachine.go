package main

import (
	"fmt"
	"strings"
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

		program   Program
		stack     *stack
		registers [16]int16

		state State

		activated map[int]bool
	}

	Program []*Gene

	runFunc func(*Machine, AST, func())

	State interface {
		// Reset resets the bot for each cycle
		// Accumulated values (such as thrust vector) are reset
		Reset()
		// Execute takes accumulated values and activates
		// the bot's systems
		Execute()
		// Returns current velocity vector X component
		X() int16
		// Returns current velocity vector Y component
		Y() int16
		// Returns current energy value, that is mass * Leonhard efficiency
		Energy() int16
		// Returns bot's ID
		ID() int16
		RemoteID(int16) int16
		Scan(int16, int16) (int16, int16)
		Thrust(int16, int16)
		Turn(int16)
		Mine(int16)
		Reproduce(int16)
		Impulse(int16)
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

func (p Program) String() string {
	var b strings.Builder
	for i, g := range p {
		b.WriteString(g.String())
		if i < len(p)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

func NewMachine() *Machine {
	m := &Machine{
		run:   runInstruction,
		stack: stackPool.Get().(*stack),

		activated: make(map[int]bool),
	}
	return m
}

func (m *Machine) Destroy() {
	stackPool.Put(m.stack)
	m.stack = nil
}

func (m *Machine) Run() {
	m.state.Reset()
	for i, g := range m.program {
		m.activated[i] = m.RunGene(g)
	}
	m.state.Execute()
}

func (m *Machine) RunGene(g *Gene) bool {
	m.pc = 0
	m.stack.Reset()
	if len(g.Evaluate) == 0 {
		return false
	}
	for {
		inst := g.Evaluate[m.pc]
		inst.Run(m, g.Evaluate)
		m.pc++
		if m.pc > len(g.Evaluate)-1 {
			break
		}
	}
	if len(*m.stack) <= 0 {
		return false
	}
	if m.stack.Pop() < 1 {
		return false
	}

	if len(g.Execute) == 0 {
		return true
	}
	m.pc = 0
	m.stack.Reset()
	for {
		inst := g.Execute[m.pc]
		inst.Run(m, g.Execute)
		m.pc++
		if m.pc > len(g.Execute)-1 {
			break
		}
	}
	return true
}

func runInstruction(m *Machine, code AST, f func()) {
	f()
}

func runInstructionDebug(m *Machine, code AST, f func()) {
	fmt.Printf("%v\n", code)
	fmt.Println("pc", m.pc)
	inst := code[m.pc]
	fmt.Printf("%s\n", inst)
	fmt.Println("exec")

	f()

	fmt.Printf("stack: %v\n\n", m.stack)
}

func runWithBreak(breakpoint int, breakFunc func(m *Machine) bool, runFunc runFunc) runFunc {
	var ran bool
	return func(m *Machine, code AST, f func()) {
		if !ran && m.pc == breakpoint {
			ran = true
			if !breakFunc(m) {
				return
			}
		}
		runFunc(m, code, f)
	}
}
