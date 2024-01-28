package main

import (
	"image"
	"math/rand"
	"sort"

	"github.com/jakecoffman/cp"
)

const (
	asteroidFrictionCoeff = 0.6
	asteroidBoundsPadding = 2.0
)

type (
	Asteroid struct {
		Bounds image.Rectangle

		*cp.Body
		*cp.Shape

		space *cp.Space
	}

	vects []cp.Vector
)

func (vs vects) Len() int           { return len(vs) }
func (vs vects) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs vects) Less(i, j int) bool { return vs[i].ToAngle() < vs[j].ToAngle() }

func NewAsteroid(sp *cp.Space, bounds image.Rectangle) *Asteroid {
	a := &Asteroid{
		space:  sp,
		Bounds: bounds,
	}
	a.Body = cp.NewBody(0, 0)
	return a
}

func (a *Asteroid) generate(seed int64) {
	rnd := rand.New(rand.NewSource(seed))

	numVerts := rnd.Intn(55) + 10

	// generate 2 lists of random x and y coordinates
	xs := make([]float64, numVerts)
	ys := make([]float64, numVerts)
	for i := 0; i < numVerts; i++ {
		xs[i] = rnd.Float64() * float64(a.Bounds.Dx())
		ys[i] = rnd.Float64() * float64(a.Bounds.Dy())
	}
	// sort them
	sort.Float64s(xs)
	sort.Float64s(ys)

	x1 := make([]cp.Vector, 0, numVerts/2)
	x2 := make([]cp.Vector, 0, numVerts/2)
	y1 := make([]cp.Vector, 0, numVerts/2)
	y2 := make([]cp.Vector, 0, numVerts/2)
	// isolate the extreme points for x and y coordinates
	// indexes 0 and len()-1
	// randomly divde them into two chains
	for i := 1; i < numVerts-1; i++ {
		if rnd.Intn(2) == 0 {
			x1 = append(x1, cp.Vector{X: xs[i]})
			y1 = append(y1, cp.Vector{Y: ys[i]})
		} else {
			x2 = append(x2, cp.Vector{X: xs[i]})
			y2 = append(y2, cp.Vector{Y: ys[i]})
		}
	}

	// combine them into vectors
	combiner := func(start, end cp.Vector, vectors []cp.Vector, neg bool) []cp.Vector {
		for i, vec := range vectors {
			tmp := vec
			if neg {
				vectors[i] = start.Sub(vec)
			} else {
				vectors[i] = vec.Sub(start)
			}
			start = tmp
		}
		if neg {
			return append(vectors, start.Sub(end))
		} else {
			return append(vectors, end.Sub(start))
		}
	}
	// combine mixed
	x1 = append(combiner(cp.Vector{X: xs[0]}, cp.Vector{X: xs[numVerts-1]}, x1, false), combiner(cp.Vector{X: xs[0]}, cp.Vector{X: xs[numVerts-1]}, x2, true)...)
	y1 = append(combiner(cp.Vector{Y: ys[0]}, cp.Vector{Y: ys[numVerts-1]}, y1, false), combiner(cp.Vector{Y: ys[0]}, cp.Vector{Y: ys[numVerts-1]}, y2, true)...)

	// randomy pair up x and y
	rnd.Shuffle(len(x1), func(i, j int) { x1[i], x1[j] = x1[j], x1[i] })
	rnd.Shuffle(len(y1), func(i, j int) { y1[i], y1[j] = y1[j], y1[i] })

	// sort by angle
	vs := make([]cp.Vector, 0, numVerts)
	for i, v := range x1 {
		vs = append(vs, v.Add(y1[i]))
	}
	sort.Sort(vects(vs))

	// build vector path and determine lower bounds
	var vect cp.Vector
	minX, minY := 0.0, 0.0
	for _, v := range vs {
		vect = vect.Add(v)
		if vect.X < minX {
			minX = vect.X
		}
		if vect.Y < minY {
			minY = vect.Y
		}
	}
	transform := cp.NewTransformTranslate(cp.Vector{X: -1 * minX, Y: -1 * minY})

	a.Shape = cp.NewPolyShape(a.Body, len(vs), vs, transform, 1)
	a.Shape.SetFriction(asteroidFrictionCoeff)
	a.Shape.Filter.Categories = SHAPE_CATEGORY_ASTEROID
	a.Shape.SetDensity(10)
	a.Body.AddShape(a.Shape)
	a.Body.AccumulateMassFromShapes()
	a.space.AddBody(a.Body)
	a.space.AddShape(a.Shape)

	// a.Path = &vector.Path{}
	// vect = vs[0]
	// pt := transform.Point(vect)
	// start := pt
	// a.Path.MoveTo(float32(pt.X), float32(pt.Y))
	//
	//	for i := 1; i < len(vs); i++ {
	//		vect = vect.Add(vs[i])
	//		pt = transform.Point(vect)
	//		a.Path.LineTo(float32(pt.X), float32(pt.Y))
	//	}
	//
	// a.Path.LineTo(float32(start.X), float32(start.Y))
	//
	// op := &ebiten.DrawTrianglesOptions{}
	// op.FillRule = ebiten.EvenOdd
	//
	// verts, _ := a.Path.AppendVerticesAndIndicesForFilling(nil, nil)
	//
	//	for i := range verts {
	//		verts[i].SrcX = 1
	//		verts[i].SrcY = 1
	//		verts[i].ColorR = 0xf0 / float32(0xff)
	//		verts[i].ColorG = 0xf0 / float32(0xff)
	//		verts[i].ColorB = 0xf0 / float32(0xff)
	//	}
}

func (a *Asteroid) Draw(g *Game) {
}
