package main

import "github.com/hajimehoshi/ebiten"

type (
	Game struct {
		scale float64

		p *Player
	}
)

func (g *Game) Update(screen *ebiten.Image) error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.p.Draw(screen)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return w, h
}
