package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
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

		space *cp.Space
		bots  []*Bot

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

func (g *Game) init() {
	b := &Bot{
		Body:    g.space.AddBody(cp.NewBody(1000000, cp.INFINITY)),
		machine: NewMachine(),
	}
	b.SetPosition(cp.Vector{X: 0, Y: 0})
	b.SetVelocity(400, 0)

	g.bots = []*Bot{b}
}

func (g *Game) Update() error {
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	g.w, g.h = w, h
	return w, h
}
