package main

import (
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
	asteroidSrc *ebiten.Image
)

func init() {
	asteroidSrc = ebiten.NewImage(1, 1)
	asteroidSrc.Fill(color.RGBA64{128, 128, 128, 255})
}

type (
	Asteroid struct {
		Bounds image.Rectangle

		Body  *cp.Body
		Shape *cp.PolyShape

		Path *vector.Path
	}
)

func NewAsteroid(bounds image.Rectangle) *Asteroid {
	a := &Asteroid{
		Bounds: bounds,
	}
	a.Body = cp.NewBody(0, 0)
	return a
}

func (a *Asteroid) generate(seed int64) {
	rnd := rand.New(rand.NewSource(seed))
	// starting point is somewhere to the right edge, around the middle
	x := float64(a.Bounds.Max.X) - asteroidBoundsPadding
	y := float64(a.Bounds.Max.Y) / 2
	// displace the y
	// maximum displacement 5% of the height
	disp := float64(a.Bounds.Size().Y) * 0.05
	disp *= rnd.Float64()
	if rnd.Intn(1) == 0 {
		disp *= -1
	}

	start := cp.Vector{X: x, Y: y}
	a.Path = &vector.Path{}
	a.Path.MoveTo(float32(start.X), float32(start.Y))

	// length
	l := rnd.Float64() * float64(a.Bounds.Dx()) * 0.3
	// angle to next vertex
	rad := rnd.Float64() * 0.5
	// add pi / 2, so we start at north
	rad += math.Pi / 2
	dispV := cp.ForAngle(rad)
	dispV.Mult(l)
	end := start.Add(dispV)
	a.Path.LineTo(float32(end.X), float32(end.Y))

	sumRad := rad
	for (math.Pi*2)-sumRad > 0.2 {
		sumRad += rnd.Float64() * 0.5
		newStart := end
		// length
		l = rnd.Float64() * float64(a.Bounds.Dx()) * 0.3
		dispV = cp.ForAngle(sumRad)
		dispV.Mult(l)
		end = newStart.Add(dispV)
		a.Path.LineTo(float32(end.X), float32(end.Y))
	}
	a.Path.LineTo(float32(start.X), float32(start.Y))
}
