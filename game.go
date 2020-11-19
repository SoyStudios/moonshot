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

const baseZoomFactor = 1.01

type (
	camera struct {
		ViewPort f64.Vec2
		Position f64.Vec2
		zoomStep int
		rotation int
	}

	Game struct {
		settings struct {
			cameraMoveSpeed float64
		}

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
		"T: %.1f, VP: %.1f, R: %d, S: %d, Z: %.1f",
		c.Position, c.ViewPort, c.rotation, c.zoomStep, c.zoomFactor(),
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
	// We want to scale and rotate around center of image / screen
	m.Translate(-c.viewportCenter()[0], -c.viewportCenter()[1])
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
	inverseMatrix.Translate(-c.Position[0], -c.Position[1])
	if inverseMatrix.IsInvertible() {
		inverseMatrix.Invert()
		return inverseMatrix.Apply(float64(posX), float64(posY))
	} else {
		// When scaling it can happend that matrix is not invertable
		return math.NaN(), math.NaN()
	}
}

func (c *camera) zoomFactor() float64 {
	return math.Pow(baseZoomFactor, float64(c.zoomStep))
}

func (c *camera) zoomTo(x, y float64) {
	op := ebiten.GeoM{}
	// magnitude
	mag := math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2))
	// unit vector
	uv := f64.Vec2{
		x / mag,
		y / mag,
	}
	op.Translate(uv[0], uv[1])
	c.Position[0], c.Position[1] = op.Apply(c.Position[0], c.Position[1])
}

func (g *Game) init() {
	g.world = ebiten.NewImage(g.w, g.h)

	b := NewBot(g.space)
	b.SetPosition(cp.Vector{X: 0, Y: 100})
	b.SetVelocity(100, 0)

	g.bots = []*Bot{b}

	b = NewBot(g.space)
	b.SetPosition(cp.Vector{X: 600, Y: 100})
	b.SetVelocity(-10, 0)
	g.bots = append(g.bots, b)

	b = NewBot(g.space)
	b.SetPosition(cp.Vector{X: 200, Y: 200})
	g.bots = append(g.bots, b)
}

func (g *Game) Update() error {
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrExit
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.camera.zoomStep--
		g.camera.zoomTo(g.camera.ScreenToWorld(ebiten.CursorPosition()))
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.camera.zoomStep++
		g.camera.zoomTo(g.camera.ScreenToWorld(ebiten.CursorPosition()))
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camera.Position[1] -= g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camera.Position[1] += g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camera.Position[0] -= g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camera.Position[0] += g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	for _, bot := range g.bots {
		op.GeoM.Reset()
		op.GeoM.Translate(-g.camera.Position[0], -g.camera.Position[1])
		op.GeoM.Scale(g.camera.zoomFactor(), g.camera.zoomFactor())
		op.GeoM.Translate(bot.Position().X, bot.Position().Y)
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
	return g.w, g.h
}
