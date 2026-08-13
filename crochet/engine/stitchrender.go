package engine

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SoyStudios/moonshot/crochet/pattern"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// Faked crochet-stitch geometry.
//
// The physics keeps one particle per stitch, which on its own renders as a flat
// lattice. To make it read as crochet, each stitch is drawn as a small bowed
// "V" — the shape a knit/crochet stitch's top loops make — sized from the
// stitch gauge and oriented by a local surface frame (row direction, column
// direction, outward normal). Rows of these Vs interlock into recognisable
// fabric, and because everything is derived from the current particle
// positions, the stitches deform with the simulation.
//
// The surface frame is where flat and round pieces differ. On a flat swatch the
// neighbours give a good frame directly (row × column → normal). On a curved
// piece worked in the round the neighbour cross product jitters, so the normal
// is taken from the smooth outward direction instead — point−centroid for a
// ball, the same with the vertical axis removed for a tube — and the row/column
// axes are rebuilt orthogonal to it. That keeps stitches on curved surfaces
// tidy instead of noisy.

// drawStitchFace renders a fabric as its field of stitch Vs, choosing a level
// of detail per stitch from how large it appears on screen.
func (e *Engine) drawStitchFace(f *pattern.Fabric) {
	e.renderer.setMaterial(f.Material)
	rad := float32(f.Radius) * 1.1
	cam := e.cam.Position

	var centroid rl.Vector3
	if f.Radial || f.Cylindrical {
		centroid = fabricCentroid(f)
	}

	for i := range f.Cells {
		c := &f.Cells[i]
		p := v(f.World.Particles[c.Node].Pos)
		u, col, n, ok := e.stitchFrame(f, c, p, centroid)
		if !ok {
			e.renderer.node(p, rad, c.Color)
			continue
		}
		lod := stitchLOD(cam, p, c.W, c.H)
		e.drawStitchV(p, u, col, n, c, rad, lod)
	}
}

// stitchFrame returns the local (row, column, normal) unit frame for a stitch.
func (e *Engine) stitchFrame(f *pattern.Fabric, c *pattern.StitchCell, p, centroid rl.Vector3) (u, col, n rl.Vector3, ok bool) {
	w := f.World

	if f.Radial || f.Cylindrical {
		out := rl.Vector3Subtract(p, centroid)
		if f.Cylindrical {
			out.Y = 0 // point straight out from the side of the tube
		}
		nn, okn := normOK(out)
		uu, oku := cellDir(w, c.Left, c.Right, c.Node)
		if !okn || !oku {
			return
		}
		colv, okc := normOK(rl.Vector3CrossProduct(nn, uu))
		if !okc {
			return
		}
		uu, _ = normOK(rl.Vector3CrossProduct(colv, nn)) // re-orthogonalise row
		return uu, colv, nn, true
	}

	// Flat piece: frame straight from the neighbours.
	uu, oku := cellDir(w, c.Left, c.Right, c.Node)
	colv, okc := cellDir(w, c.Down, c.Up, c.Node)
	if !oku || !okc {
		return
	}
	nn, okn := normOK(rl.Vector3CrossProduct(uu, colv))
	if !okn {
		return
	}
	return uu, colv, nn, true
}

// drawStitchV draws the bowed V for a stitch given its surface frame and LOD.
func (e *Engine) drawStitchV(p, u, col, n rl.Vector3, c *pattern.StitchCell, rad float32, lod int) {
	fw := float32(c.W)
	fh := float32(c.H)

	// The V spans (nearly) the full stitch height so a row's tips meet the next
	// row's bottom vertices, and a bit over half-width so neighbours overlap —
	// giving continuous, interlocking columns rather than isolated chevrons.
	hw := rl.Vector3Scale(u, fw*0.58)
	up := rl.Vector3Scale(col, fh*0.5)
	down := rl.Vector3Scale(col, fh*0.5)
	bulge := rl.Vector3Scale(n, (fw+fh)*0.11)

	bottom := rl.Vector3Subtract(p, down)
	topL := rl.Vector3Add(rl.Vector3Subtract(p, hw), up)
	topR := rl.Vector3Add(rl.Vector3Add(p, hw), up)

	switch lod {
	case 0: // full: two bowed arms with rounded tips
		e.arc(bottom, topL, bulge, rad, c.Color)
		e.arc(bottom, topR, bulge, rad, c.Color)
		e.renderer.node(bottom, rad, c.Color) // cap the shared base joint
	case 1: // straight V: one cylinder per arm
		e.renderer.segment(bottom, topL, rad, c.Color)
		e.renderer.segment(bottom, topR, rad, c.Color)
		e.renderer.node(bottom, rad, c.Color)
	default: // far: a single bar standing in for the whole stitch
		mid := rl.Vector3Lerp(topL, topR, 0.5)
		e.renderer.segment(bottom, mid, rad, c.Color)
	}
}

// arc draws a yarn strand from a to b, bowed by `bulge` at its midpoint, as a
// short chain of lit cylinders with a rounded tip.
func (e *Engine) arc(a, b, bulge rl.Vector3, rad float32, col yarn.Color) {
	const steps = 4
	ctrl := rl.Vector3Add(rl.Vector3Scale(rl.Vector3Add(a, b), 0.5), bulge)
	prev := a
	for i := 1; i <= steps; i++ {
		t := float32(i) / float32(steps)
		p := qBezier(a, ctrl, b, t)
		e.renderer.segment(prev, p, rad, col)
		prev = p
	}
	e.renderer.node(b, rad, col)
}

func qBezier(a, ctrl, b rl.Vector3, t float32) rl.Vector3 {
	return rl.Vector3Lerp(rl.Vector3Lerp(a, ctrl, t), rl.Vector3Lerp(ctrl, b, t), t)
}

// stitchLOD chooses 0 (full bowed V), 1 (straight V) or 2 (single bar) from the
// stitch's apparent size — its world size over its distance to the camera.
func stitchLOD(cam, p rl.Vector3, w, h float64) int {
	dx, dy, dz := float64(cam.X-p.X), float64(cam.Y-p.Y), float64(cam.Z-p.Z)
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if dist < 1e-3 {
		return 0
	}
	apparent := (w + h) * 0.5 / dist
	switch {
	case apparent > 0.03:
		return 0
	case apparent > 0.012:
		return 1
	default:
		return 2
	}
}

// fabricCentroid averages the fabric's stitch positions.
func fabricCentroid(f *pattern.Fabric) rl.Vector3 {
	var sx, sy, sz float64
	for i := range f.Cells {
		p := f.World.Particles[f.Cells[i].Node].Pos
		sx += p.X
		sy += p.Y
		sz += p.Z
	}
	n := len(f.Cells)
	if n == 0 {
		return rl.Vector3{}
	}
	inv := 1 / float64(n)
	return rl.NewVector3(float32(sx*inv), float32(sy*inv), float32(sz*inv))
}

// cellDir returns the unit direction across a stitch from one neighbour to the
// other, falling back to whichever single neighbour exists.
func cellDir(w *physics.World, a, b, node int) (rl.Vector3, bool) {
	switch {
	case a >= 0 && b >= 0:
		return normOK(rl.Vector3Subtract(v(w.Particles[b].Pos), v(w.Particles[a].Pos)))
	case b >= 0:
		return normOK(rl.Vector3Subtract(v(w.Particles[b].Pos), v(w.Particles[node].Pos)))
	case a >= 0:
		return normOK(rl.Vector3Subtract(v(w.Particles[node].Pos), v(w.Particles[a].Pos)))
	}
	return rl.Vector3{}, false
}

func normOK(x rl.Vector3) (rl.Vector3, bool) {
	l := float32(math.Sqrt(float64(x.X*x.X + x.Y*x.Y + x.Z*x.Z)))
	if l < 1e-6 {
		return x, false
	}
	return rl.Vector3Scale(x, 1/l), true
}
