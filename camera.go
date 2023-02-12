package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

// The camera has zoom steps, the factor by which each step is
// multiplied
const baseZoomFactor = 1.01

type (
	camera struct {
		// ViewPort is the size of the viewport width * height
		ViewPort cp.Vector
		// Position of the camera in the world
		Position cp.Vector
		zoomStep int
		rotation int
	}
)

func (c *camera) String() string {
	return fmt.Sprintf(
		"P: %.1f,%.1f, VP: %.1f,%.1f, R: %d, S: %d, Z: %.1f",
		c.Position.X, c.Position.Y,
		c.ViewPort.X, c.ViewPort.Y,
		c.rotation,
		c.zoomStep, c.zoomFactor(),
	)
}

func (c *camera) viewportCenter() cp.Vector {
	return cp.Vector{
		X: c.ViewPort.X * 0.5,
		Y: c.ViewPort.Y * 0.5,
	}
}

// worldObjectMatrix returns a matrix used to place an object
// onto the world on coordinates x, y
// relative to the camera
func (c *camera) worldObjectMatrix(x, y float64) ebiten.GeoM {
	g := ebiten.GeoM{}
	g.Translate(x, y)
	return c.getTransform(g)
}

func (c *camera) getTransform(g ebiten.GeoM) ebiten.GeoM {
	g.Translate(-c.Position.X, -c.Position.Y)
	g.Translate(-c.viewportCenter().X, -c.viewportCenter().Y)
	g.Scale(c.zoomFactor(), c.zoomFactor())
	g.Translate(c.viewportCenter().X, c.viewportCenter().Y)
	return g
}

func (c *camera) Render(world, screen *ebiten.Image) {
	screen.DrawImage(world, &ebiten.DrawImageOptions{})
}

// WorldToScreen translates world coordinates (such as positions of bots
// etc.) into screen coordinates for rendering onto the world plane
func (c *camera) WorldToScreen(x, y float64) (float64, float64) {
	vec := cp.Vector{X: x, Y: y}
	vec = vec.Add(c.Position.Neg())
	return vec.X, vec.Y
}

func (c *camera) ScreenToWorld(posX, posY int) (float64, float64) {
	inverseMatrix := c.worldObjectMatrix(0, 0)
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
	to := cp.Vector{X: x, Y: y}
	to = to.Clamp(1)
	to = to.Mult(1 / c.zoomFactor())
	c.Position = c.Position.Add(to)
}
