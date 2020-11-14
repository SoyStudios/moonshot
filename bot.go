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

func (b *Bot) Draw(sprite *ebiten.Image, world *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Reset()
	op.GeoM.Translate(b.Position().X, b.Position().Y)
	world.DrawImage(sprite, op)
}
