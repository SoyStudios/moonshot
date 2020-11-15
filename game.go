package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

		w, h int

		p *Player
	}
)

func (c *camera) String() string {
	return fmt.Sprintf(
		"T: %.1f, VP: %.1f, R: %d, S: %d",
		c.Position, c.ViewPort, c.rotation, c.zoomFactor,
	)
}

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

func (g *Game) init() {
	b := &Bot{
		Body:    g.space.AddBody(cp.NewBody(1000000, cp.INFINITY)),
		machine: NewMachine(),
	}
	b.SetPosition(cp.Vector{X: 100, Y: 100})
	b.SetVelocity(100, 0)

	b.shape = cp.NewCircle(b.Body, 0.95, cp.Vector{})
	b.shape.SetElasticity(0)
	b.shape.SetFriction(0)
	g.space.AddShape(b.shape)

	g.bots = []*Bot{b}
}

func (g *Game) Update() error {
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrExit
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	s := ebiten.DeviceScaleFactor()
	g.world = ebiten.NewImage(windowWidth, windowHeight)
	g.world.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	for _, bot := range g.bots {
		op.GeoM.Reset()
		op.GeoM.Translate(-g.camera.Position[0], -g.camera.Position[1])
		op.GeoM.Translate(bot.Position().X, bot.Position().Y)
		op.GeoM.Scale(s, s)
		op.GeoM.Scale(1/1.01*float64(g.camera.zoomFactor), 1/1.01*float64(g.camera.zoomFactor))
		g.world.DrawImage(g.assets.bot, op)
	}

	g.camera.Render(g.world, screen)

	worldX, worldY := g.camera.ScreenToWorld(ebiten.CursorPosition())
	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf("TPS: %0.2f\n", ebiten.CurrentTPS()),
	)

	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("%s\nCursor World Pos: %.2f,%.2f",
			g.camera.String(),
			worldX, worldY,
		),
		0, g.h-48,
	)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return windowWidth, windowHeight
}
