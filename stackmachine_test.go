package main

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type StateMock struct {
	mock.Mock
}

func (s *StateMock) Reset() {
	s.Called()
}

func (s *StateMock) X() int16 {
	args := s.Called()
	return args.Get(0).(int16)
}

func (s *StateMock) Y() int16 {
	args := s.Called()
	return args.Get(0).(int16)
}

func (s *StateMock) Energy() int16 {
	args := s.Called()
	return args.Get(0).(int16)
}

func (s *StateMock) ID() int16 {
	args := s.Called()
	return args.Get(0).(int16)
}

func (s *StateMock) RemoteID(a int16) int16 {
	args := s.Called(a)
	return args.Get(0).(int16)
}

func (s *StateMock) Scan(x, y int16) (int16, int16) {
	args := s.Called(x, y)
	return args.Get(0).(int16), args.Get(1).(int16)
}

func (s *StateMock) Thrust(a int16) {
	s.Called(a)
}

func (s *StateMock) Turn(x, y int16) {
	s.Called(x, y)
}

func (s *StateMock) Mine(a int16) {
	s.Called(a)
}

func (s *StateMock) Reproduce(a int16) {
	s.Called(a)
}

func TestSimpleMachine(t *testing.T) {
	program := []*Gene{
		&Gene{
			Evaluate: TranslateProgram([]Token{
				RDX,
				PSH, CON, 0,
				GEQ,
			}),
			Execute: TranslateProgram([]Token{
				PSH, CON, 12,
				THR,
			}),
		},
	}
	m := NewMachine()
	m.run = runInstruction
	m.program = program
	stateMock := &StateMock{}
	m.state = stateMock

	stateMock.On("Reset")
	stateMock.On("X").Return(int16(4)).Once()
	stateMock.On("Thrust", int16(12)).Once()

	m.Run()
	if !stateMock.AssertExpectations(t) {
		return
	}

	stateMock.On("X").Return(int16(-1))
	m.Run()
	if !stateMock.AssertExpectations(t) {
		return
	}

}

func BenchmarkSimpleMachine(b *testing.B) {
	program := []*Gene{
		&Gene{
			Evaluate: TranslateProgram([]Token{
				RDX,
				PSH, CON, 0,
				GEQ,
			}),
			Execute: TranslateProgram([]Token{
				PSH, CON, 12,
				THR,
			}),
		},
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := NewMachine()
			m.program = program
			stateMock := &StateMock{}
			m.state = stateMock

			stateMock.On("X").Return(int16(4))
			stateMock.On("Thrust", int16(12))

			m.Run()
		}
	})
}
