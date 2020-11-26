package main

import (
	"github.com/stretchr/testify/mock"
)

type StateMock struct {
	mock.Mock
}

func (s *StateMock) Reset() {
	s.Called()
}

func (s *StateMock) Execute() {
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

func (s *StateMock) Thrust(x, y int16) {
	s.Called(x, y)
}

func (s *StateMock) Turn(a int16) {
	s.Called(a)
}

func (s *StateMock) Mine(a int16) {
	s.Called(a)
}

func (s *StateMock) Reproduce(a int16) {
	s.Called(a)
}
