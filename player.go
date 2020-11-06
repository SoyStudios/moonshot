package main

import "github.com/hajimehoshi/ebiten"

type (
	Player struct {
		x, y, angle float64
	}
)

func (p *Player) Draw(screen *ebiten.Image) {
}
