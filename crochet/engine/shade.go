package engine

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// The renderer draws yarn as immediate-mode cylinders, which raylib shades
// flat (unlit). To give the yarn depth and let materials read differently, we
// compute a cheap per-segment shade on the CPU: a directional "sun" plus a
// specular sheen. It is not physically exact, but it makes tubes look round and
// distinguishes matte wool from glossy acrylic.

// lightDir is the direction the key light travels (from upper-left, downward).
var lightDir = math3.V(-0.45, -1.0, -0.35).Normalize()

// shadeSegment returns the lit colour of a yarn segment from a→b viewed from
// cam, given its base colour and material.
func shadeSegment(base yarn.Color, mat yarn.Material, a, b, cam math3.Vec3) rl.Color {
	axis := b.Sub(a)
	if axis.LenSq() == 0 {
		return shadePoint(base, mat)
	}
	axis = axis.Normalize()
	toLight := lightDir.Neg()

	// A cylinder presents its broadside to the light when its axis is
	// perpendicular to the light; use that as the diffuse term.
	al := axis.Dot(toLight)
	lambert := math.Sqrt(math.Max(0, 1-al*al))
	d := mat.Ambient + (1-mat.Ambient)*lambert

	// Specular sheen: bright where the half-vector is broadside to the axis.
	mid := a.Add(b).Scale(0.5)
	view := cam.Sub(mid).Normalize()
	half := toLight.Add(view).Normalize()
	ah := axis.Dot(half)
	spec := math.Pow(math.Max(0, 1-ah*ah), 8) * mat.Sheen

	return rl.NewColor(
		chan8(float64(base.R)*d+255*spec),
		chan8(float64(base.G)*d+255*spec),
		chan8(float64(base.B)*d+255*spec),
		base.A,
	)
}

// shadePoint returns a mildly-lit colour for round elements (nodes, poles),
// which have no single axis to shade by.
func shadePoint(base yarn.Color, mat yarn.Material) rl.Color {
	d := mat.Ambient + (1-mat.Ambient)*0.8
	return rl.NewColor(
		chan8(float64(base.R)*d),
		chan8(float64(base.G)*d),
		chan8(float64(base.B)*d),
		base.A,
	)
}

func chan8(x float64) uint8 {
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return uint8(x)
}

func darker(c yarn.Color, f float64) yarn.Color {
	return yarn.Color{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: c.A,
	}
}
