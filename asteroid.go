package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
)

const (
	asteroidFrictionCoeff = 0.6
	asteroidBoundsPadding = 2.0
)

var (
	asteroidBase *ebiten.Image
	asteroidSrc  *ebiten.Image
)

func init() {
	asteroidBase = ebiten.NewImage(3, 3)
	asteroidBase.Fill(color.White)
	asteroidSrc = asteroidBase.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
}

type (
	Asteroid struct {
		Bounds image.Rectangle

		*cp.Body
		*cp.Shape

		Path *vector.Path
		Img  *ebiten.Image
	}
)

func NewAsteroid(bounds image.Rectangle) *Asteroid {
	a := &Asteroid{
		Bounds: bounds,
	}
	a.Img = ebiten.NewImageWithOptions(a.Bounds, nil)
	a.Body = cp.NewBody(0, 0)
	return a
}

func (a *Asteroid) generate(seed int64) {
	rnd := rand.New(rand.NewSource(seed))
	x := float64(a.Bounds.Max.X) / 2
	y := float64(a.Bounds.Max.Y) / 2

	center := cp.Vector{X: x, Y: y}

	rad := 0.0
	// length
	l := (float64(a.Bounds.Dx()) / 2) - asteroidBoundsPadding
	dispV := cp.ForAngle(rad)
	start := center.Add(dispV.Mult(l))

	a.Path = &vector.Path{}
	a.Path.MoveTo(float32(start.X), float32(start.Y))

	sumRad := rad
	for (math.Pi*2)-sumRad > 0.2 {
		sumRad += rnd.Float64() * 0.5
		dispV = cp.ForAngle(sumRad)
		point := center.Add(dispV.Mult(l))
		a.Path.LineTo(float32(point.X), float32(point.Y))
		infoLog.Log("msg", "line",
			"to", fmt.Sprintf("%.0f, %.0f", point.X, point.Y),
		)
	}
	a.Path.LineTo(float32(start.X), float32(start.Y))

	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.EvenOdd

	a.Img.Clear()
	verts, indexes := a.Path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range verts {
		verts[i].SrcX = 1
		verts[i].SrcY = 1
		verts[i].ColorR = 0xf0 / float32(0xff)
		verts[i].ColorG = 0xf0 / float32(0xff)
		verts[i].ColorB = 0xf0 / float32(0xff)
	}
	a.Img.DrawTriangles(verts, indexes, asteroidSrc, op)
}

func (a *Asteroid) Draw(g *Game) {
	op := &ebiten.DrawImageOptions{}
	dx, dy := float64(a.Bounds.Dx())/2, float64(a.Bounds.Dy())/2
	op.GeoM = g.camera.worldObjectMatrix(
		a.Position().X-dx,
		a.Position().Y-dy,
	)
	g.world.DrawImage(a.Img, op)
}
