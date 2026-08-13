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
// "V" — the shape the top loops of a knit/crochet stitch make — sized from the
// stitch gauge and oriented by the *live* fabric surface. The surface frame at a
// stitch comes from its neighbours: the row direction (left→right) and column
// direction (down→up) span the fabric, and their cross product is the outward
// normal the V bows along. Rows of these Vs interlock into recognisable
// stockinette-like fabric, and because everything is derived from the current
// particle positions, the stitches deform with the simulation.

// drawStitchFace renders a fabric as its field of stitch Vs, picking a
// level of detail per stitch from how large it appears on screen.
func (e *Engine) drawStitchFace(f *pattern.Fabric) {
	e.renderer.setMaterial(f.Material)
	rad := float32(f.Radius) * 1.1
	cam := e.cam.Position
	for i := range f.Cells {
		c := &f.Cells[i]
		lod := stitchLOD(cam, v(f.World.Particles[c.Node].Pos), c.W, c.H)
		e.drawStitch(f.World, c, rad, lod)
	}
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

// drawStitch draws a single stitch cell as a bowed V at the given LOD.
func (e *Engine) drawStitch(w *physics.World, c *pattern.StitchCell, rad float32, lod int) {
	p := v(w.Particles[c.Node].Pos)
	u, okU := cellDir(w, c.Left, c.Right, c.Node)
	vv, okV := cellDir(w, c.Down, c.Up, c.Node)
	if !okU || !okV {
		// Edge or degenerate stitch: draw a bead so the fabric has no holes.
		e.renderer.node(p, rad, c.Color)
		return
	}

	n, okN := normOK(rl.Vector3CrossProduct(u, vv))
	if !okN {
		e.renderer.node(p, rad, c.Color)
		return
	}
	fw := float32(c.W)
	fh := float32(c.H)

	// The V spans (nearly) the full stitch height so a row's tips meet the next
	// row's bottom vertices, and a bit over half-width so neighbours overlap —
	// giving continuous, interlocking columns rather than isolated chevrons.
	hw := rl.Vector3Scale(u, fw*0.58)         // half stitch width along the row
	up := rl.Vector3Scale(vv, fh*0.5)         // toward the two top tips
	down := rl.Vector3Scale(vv, fh*0.5)       // toward the bottom vertex
	bulge := rl.Vector3Scale(n, (fw+fh)*0.12) // pop the V out of the surface

	bottom := rl.Vector3Subtract(p, down)
	topL := rl.Vector3Add(rl.Vector3Subtract(p, hw), up)
	topR := rl.Vector3Add(rl.Vector3Add(p, hw), up)

	switch lod {
	case 0: // full: two bowed arms with rounded tips
		e.arc(bottom, topL, bulge, rad, c.Color)
		e.arc(bottom, topR, bulge, rad, c.Color)
	case 1: // straight V: one cylinder per arm, no bezier, no tips
		e.renderer.segment(bottom, topL, rad, c.Color)
		e.renderer.segment(bottom, topR, rad, c.Color)
	default: // far: a single bar standing in for the whole stitch
		mid := rl.Vector3Lerp(topL, topR, 0.5)
		e.renderer.segment(bottom, mid, rad, c.Color)
	}
}

// arc draws a yarn strand from a to b, bowed by `bulge` at its midpoint, as a
// short chain of lit cylinders with a rounded tip.
func (e *Engine) arc(a, b, bulge rl.Vector3, rad float32, col yarn.Color) {
	const steps = 3
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
