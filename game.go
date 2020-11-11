package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/math/f64"
)

type (
	camera struct {
		ViewPort   f64.Vec2
		Position   f64.Vec2
		zoomFactor int
	}

	Game struct {
		assets *assets

		camera *camera

		// width, height
		w, h int
		p    *Player
	}
)

func (c *camera) viewportCenter() f64.Vec2 {
	return f64.Vec2{
		c.ViewPort[0] * 0.5,
		c.ViewPort[1] * 0.5,
	}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	g.w, g.h = w, h
	return w, h
}
