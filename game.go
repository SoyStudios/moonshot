package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"golang.org/x/image/math/f64"
)

type (
	camera struct {
		ViewPort   f64.Vec2
		Position   f64.Vec2
		zoomFactor int
		rotation   int
	}

	Game struct {
		assets *assets

		camera *camera

		tps int

		space *cp.Space
		bots  []*Bot

		world *ebiten.Image

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
func (c *camera) worldMatrix() ebiten.GeoM {
	m := ebiten.GeoM{}
	m.Translate(-c.Position[0], -c.Position[1])
	// We want to scale and rotate around center of image / screen
	m.Translate(-c.viewportCenter()[0], -c.viewportCenter()[1])
	m.Scale(
		math.Pow(1.01, float64(c.zoomFactor)),
		math.Pow(1.01, float64(c.zoomFactor)),
	)
	m.Rotate(float64(c.rotation) * 2 * math.Pi / 360)
	m.Translate(c.viewportCenter()[0], c.viewportCenter()[1])
	return m
}

func (c *camera) Render(world, screen *ebiten.Image) {
	screen.DrawImage(world, &ebiten.DrawImageOptions{
		GeoM: c.worldMatrix(),
	})
}

func (c *camera) ScreenToWorld(posX, posY int) (float64, float64) {
	inverseMatrix := c.worldMatrix()
	if inverseMatrix.IsInvertible() {
		inverseMatrix.Invert()
		return inverseMatrix.Apply(float64(posX), float64(posY))
	} else {
		// When scaling it can happend that matrix is not invertable
		return math.NaN(), math.NaN()
	}
}

func (g *Game) init(worldWidth, worldHeight int) {
	g.world = ebiten.NewImage(int(g.camera.ViewPort[0]/2)+1, int(g.camera.ViewPort[1]/2)+1)

	b := &Bot{
		Body:    g.space.AddBody(cp.NewBody(1000000, cp.INFINITY)),
		machine: NewMachine(),
	}
	b.SetPosition(cp.Vector{X: 0, Y: 0})
	b.SetVelocity(400, 0)

	b.shape = cp.NewCircle(b.Body, 0.95, cp.Vector{})
	b.shape.SetElasticity(0)
	b.shape.SetFriction(0)
	g.space.AddShape(b.shape)

	g.bots = []*Bot{b}
}

func (g *Game) Update() error {
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Fill(color.White)

	op := &ebiten.DrawImageOptions{}
	for _, bot := range g.bots {
		op.GeoM.Reset()
		op.GeoM.Translate(-g.camera.Position[0], -g.camera.Position[1])
		op.GeoM.Translate(-g.camera.viewportCenter()[0], -g.camera.viewportCenter()[1])
		op.GeoM.Translate(bot.Position().X, bot.Position().Y)
		g.world.DrawImage(g.assets.bot, op)
	}

	g.camera.Render(g.world, screen)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	g.w, g.h = w, h
	return w, h
}
