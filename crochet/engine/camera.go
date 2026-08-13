package engine

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SoyStudios/moonshot/crochet/math3"
)

// orbitCam is a simple turntable camera: it looks at a target from a spherical
// offset. The left mouse button orbits, the wheel zooms, and the right mouse
// button pans the target in the camera's screen plane.
type orbitCam struct {
	target             math3.Vec3
	azimuth, elevation float64 // radians
	distance           float64
}

func newOrbit(target math3.Vec3) orbitCam {
	return orbitCam{
		target:    target,
		azimuth:   0.6,
		elevation: 0.5,
		distance:  14,
	}
}

func (o *orbitCam) handleInput() {
	const (
		rotSpeed  = 0.005
		panSpeed  = 0.01
		zoomSpeed = 1.2
		minPitch  = -1.4
		maxPitch  = 1.4
		minDist   = 1.5
		maxDist   = 80
	)

	if wheel := rl.GetMouseWheelMove(); wheel != 0 {
		o.distance -= float64(wheel) * zoomSpeed
		o.distance = clamp(o.distance, minDist, maxDist)
	}

	d := rl.GetMouseDelta()

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		o.azimuth -= float64(d.X) * rotSpeed
		o.elevation += float64(d.Y) * rotSpeed
		o.elevation = clamp(o.elevation, minPitch, maxPitch)
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		right, up := o.basis()
		scale := panSpeed * o.distance / 14
		o.target = o.target.
			Add(right.Scale(-float64(d.X) * scale)).
			Add(up.Scale(float64(d.Y) * scale))
	}
}

// position returns the world-space camera position for the current spherical
// parameters.
func (o *orbitCam) position() math3.Vec3 {
	ce := math.Cos(o.elevation)
	return o.target.Add(math3.V(
		o.distance*ce*math.Sin(o.azimuth),
		o.distance*math.Sin(o.elevation),
		o.distance*ce*math.Cos(o.azimuth),
	))
}

// basis returns the camera's right and up vectors in world space (used for
// panning).
func (o *orbitCam) basis() (right, up math3.Vec3) {
	forward := o.target.Sub(o.position()).Normalize()
	worldUp := math3.V(0, 1, 0)
	right = forward.Cross(worldUp).Normalize()
	up = right.Cross(forward).Normalize()
	return right, up
}

func (o *orbitCam) apply(cam *rl.Camera3D) {
	cam.Position = v(o.position())
	cam.Target = v(o.target)
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
