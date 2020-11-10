package main

import "github.com/hajimehoshi/ebiten/v2"

type (
	Game struct {
		viewport struct {
			scale float64
			x, y  float64
		}

		// width, height
		w, h int
		p    *Player
	}
)

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	g.w, g.h = w, h
	return w, h
}
