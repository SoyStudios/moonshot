package main

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type StateMock struct {
	mock.Mock
}

func (s *StateMock) X() int16 {
	args := s.Called("X")
	return args.Get(0).(int16)
}

func (s *StateMock) Y() int16 {
	args := s.Called("Y")
	return args.Get(0).(int16)
}

func (s *StateMock) Angle() int16 {
	args := s.Called("Angle")
	return args.Get(0).(int16)
}

func (s *StateMock) Energy() int16 {
	args := s.Called("Energy")
	return args.Get(0).(int16)
}

func (s *StateMock) ID() int16 {
	args := s.Called("ID")
	return args.Get(0).(int16)
}

func (s *StateMock) RemoteID(a int16) int16 {
	args := s.Called("RemoteID", a)
	return args.Get(0).(int16)
}

func (s *StateMock) Scan(a int16) int16 {
	args := s.Called("Scan", a)
	return args.Get(0).(int16)
}

func (s *StateMock) Thrust(a int16) {
	s.Called("Thrust", a)
}

func (s *StateMock) Turn(a int16) {
	s.Called("Turn", a)
}

func (s *StateMock) Mine(a int16) {
	s.Called("Mine", a)
}

func (s *StateMock) Reproduce(a int16) {
	s.Called("Reproduce", a)
}

func TestSimpleMachine(t *testing.T) {
	program := Program{
		Evaluate: TranslateProgram([]Token{
			RDX,
			PSH, CON, 0,
			GEQ,
		}),
		Execute: TranslateProgram([]Token{
			PSH, CON, 12,
			THR,
		}),
	}
	m := NewMachine()
	m.program = program
	stateMock := &StateMock{}
	m.state = stateMock
}
