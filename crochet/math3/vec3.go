// Package math3 provides a small, dependency-free 3D vector type used by the
// crochet engine's physics layer. Keeping it independent of raylib lets the
// simulation be unit-tested in a headless environment (no cgo, no window).
package math3

import "math"

// Vec3 is a 3D vector using float64 for simulation precision. The renderer
// converts to float32 only at draw time.
type Vec3 struct {
	X, Y, Z float64
}

// V is a convenience constructor.
func V(x, y, z float64) Vec3 { return Vec3{x, y, z} }

// Zero is the additive identity.
var Zero = Vec3{}

// Add returns a + b.
func (a Vec3) Add(b Vec3) Vec3 { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }

// Sub returns a - b.
func (a Vec3) Sub(b Vec3) Vec3 { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }

// Scale returns a * s.
func (a Vec3) Scale(s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }

// Neg returns -a.
func (a Vec3) Neg() Vec3 { return Vec3{-a.X, -a.Y, -a.Z} }

// Dot returns the dot product a·b.
func (a Vec3) Dot(b Vec3) float64 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }

// Cross returns the cross product a×b.
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

// LenSq returns the squared length (cheaper than Len; avoids the sqrt).
func (a Vec3) LenSq() float64 { return a.Dot(a) }

// Len returns the Euclidean length.
func (a Vec3) Len() float64 { return math.Sqrt(a.LenSq()) }

// Distance returns the distance between a and b.
func (a Vec3) Distance(b Vec3) float64 { return a.Sub(b).Len() }

// Normalize returns a unit vector in the direction of a. A zero vector is
// returned unchanged to avoid producing NaNs.
func (a Vec3) Normalize() Vec3 {
	l := a.Len()
	if l == 0 {
		return a
	}
	return a.Scale(1 / l)
}

// Lerp linearly interpolates from a to b by t in [0,1].
func Lerp(a, b Vec3, t float64) Vec3 {
	return Vec3{
		a.X + (b.X-a.X)*t,
		a.Y + (b.Y-a.Y)*t,
		a.Z + (b.Z-a.Z)*t,
	}
}
