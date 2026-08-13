package physics

import "github.com/SoyStudios/moonshot/crochet/math3"

// PressureConstraint models the stuffing inside a closed crochet shape
// (amigurumi). Rather than simulating a gas volume exactly, it pushes every
// member particle radially outward from the shape's current centroid by a
// small amount each solve. The surrounding yarn constraints resist, so the
// shape settles at an inflated equilibrium — a soft, stuffed look.
//
// It satisfies the Constraint interface, so it is relaxed alongside the yarn
// each iteration.
type PressureConstraint struct {
	Members  []*Particle
	Strength float64 // outward displacement per solve, in world units
}

// NewPressure builds a stuffing constraint over the given particles.
func NewPressure(members []*Particle, strength float64) *PressureConstraint {
	return &PressureConstraint{Members: members, Strength: strength}
}

// Solve pushes each member outward from the members' centroid.
func (c *PressureConstraint) Solve() {
	if c.Strength == 0 || len(c.Members) == 0 {
		return
	}
	var ctr math3.Vec3
	for _, m := range c.Members {
		ctr = ctr.Add(m.Pos)
	}
	ctr = ctr.Scale(1 / float64(len(c.Members)))

	for _, m := range c.Members {
		if m.InvMass == 0 {
			continue
		}
		dir := m.Pos.Sub(ctr)
		d := dir.Len()
		if d == 0 {
			continue
		}
		m.Pos = m.Pos.Add(dir.Scale(c.Strength / d))
	}
}
