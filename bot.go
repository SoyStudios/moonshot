package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type (
	Bot struct {
		*cp.Body

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

func NewBot(sp *cp.Space) *Bot {
	b := &Bot{
		Body:    sp.AddBody(cp.NewBody(100, cp.INFINITY)),
		machine: NewMachine(),
	}
	b.shape = cp.NewCircle(b.Body, 8, cp.Vector{})
	b.shape.SetElasticity(0)
	b.shape.SetFriction(0)
	sp.AddShape(b.shape)

	return b
}

func (b *Bot) Draw(sprite *ebiten.Image, world *ebiten.Image) {
}
