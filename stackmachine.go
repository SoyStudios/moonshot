package main

import "sync"

type (
	stack []int16

	Machine struct {
		pc int // program counter
		i  int // instruction counter

		program   []int16
		stack     stack
		callStack stack

		state State
	}

	State interface {
		X() float64
		Y() float64
		Angle() float64
		Health() float64
		Energy() float64
		Thrust(int16)
	}

	ReadX struct{}
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
		callStack: stackPool.Get().(stack),
		stack:     stackPool.Get().(stack),
	}
	return m
}
