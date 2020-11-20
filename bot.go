package main

import (
	"math"

	"github.com/jakecoffman/cp"
)

type (
	Bot struct {
		*cp.Body

		space *cp.Space

		id int16

		leonhardEfficiency func() float64
		// thrustStep translates a given thrust step value
		// to a force
		thrustStep func(int16) float64

		shape   *cp.Shape
		machine *Machine
	}
)

func BotRunner(bc <-chan *Bot, done chan struct{}) {
	for {
		b, ok := <-bc
		if !ok {
			break
		}
		b.machine.Run()
		done <- struct{}{}
	}
}

func NewBot(sp *cp.Space, id int16) *Bot {
	b := &Bot{
		Body: sp.AddBody(cp.NewBody(100, cp.INFINITY)),

		space: sp,

		id: id,

		leonhardEfficiency: func() float64 {
			return .65
		},
		thrustStep: func(step int16) float64 {
			switch true {
			case step < 100:
				return 80
			case step < 200:
				return 140
			default:
				return 200
			}
		},
		machine: NewMachine(),
	}
	b.machine.state = b
	b.shape = cp.NewCircle(b.Body, 8, cp.Vector{})
	b.shape.SetElasticity(0)
	b.shape.SetFriction(0)
	sp.AddShape(b.shape)

	return b
}

func (b *Bot) X() int16 {
	return int16(math.Round(b.Velocity().X))
}

func (b *Bot) Y() int16 {
	return int16(math.Round(b.Velocity().Y))
}

func (b *Bot) Energy() int16 {
	return int16(math.Round(b.Mass() * b.leonhardEfficiency()))
}

func (b *Bot) ID() int16 {
	return b.id
}

func (b *Bot) RemoteID(int16) int16 {
	return 1
}

func (b *Bot) Scan(x, y int16) (int16, int16) {
	return 0, 0
}

func (b *Bot) Thrust(step int16) {
	dir := cp.ForAngle(b.Angle())
	dir = dir.Neg()
	dir = dir.Mult(b.thrustStep(step))
	b.ApplyImpulseAtLocalPoint(
		dir,
		cp.Vector{X: 0, Y: 0},
	)
}

func (b *Bot) Turn(x, y int16) {
	v := cp.Vector{X: float64(x), Y: float64(y)}
	a := v.ToAngle()
	b.SetAngularVelocity(a)
}

func (b *Bot) Mine(strength int16) {
}

func (b *Bot) Reproduce(energy int16) {
}
