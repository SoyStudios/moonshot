package physics

// Constraint is anything that can nudge particles back toward a valid
// configuration. Constraints are solved iteratively each step; more iterations
// make the assembly stiffer and less stretchy.
type Constraint interface {
	Solve()
}

// DistanceConstraint keeps two particles at a target Rest distance. It is the
// fundamental building block of yarn: chaining many of them along a line makes
// a rope, and cross-linking them makes fabric.
//
// The same type doubles as a bending/stiffness constraint — link particle i to
// i+2 with a lower Stiffness to resist the yarn folding sharply.
type DistanceConstraint struct {
	A, B *Particle
	Rest float64 // target separation

	// Stiffness in [0,1] scales the positional correction applied per
	// iteration. 1 is rigid; smaller values are springy/stretchy. Because
	// corrections compound across solver iterations, the effective stiffness
	// is higher than this single-pass value.
	Stiffness float64

	// Tear, when > 0, is a stretch ratio at which the constraint snaps. If the
	// current distance exceeds Rest*Tear the constraint breaks and stops
	// solving, modelling yarn ripping apart.
	Tear   float64
	broken bool
}

// NewDistanceConstraint links a and b at their current separation.
func NewDistanceConstraint(a, b *Particle, stiffness float64) *DistanceConstraint {
	return &DistanceConstraint{
		A:         a,
		B:         b,
		Rest:      a.Pos.Distance(b.Pos),
		Stiffness: stiffness,
	}
}

// Broken reports whether a tearable constraint has snapped.
func (c *DistanceConstraint) Broken() bool { return c.broken }

// Solve applies one relaxation pass, moving A and B toward the rest distance in
// proportion to their inverse masses (so heavier particles move less, and
// pinned particles do not move at all).
func (c *DistanceConstraint) Solve() {
	if c.broken {
		return
	}
	invSum := c.A.InvMass + c.B.InvMass
	if invSum == 0 {
		return // both pinned; nothing to do
	}

	delta := c.B.Pos.Sub(c.A.Pos)
	dist := delta.Len()
	if dist == 0 {
		return // coincident; direction undefined
	}

	if c.Tear > 0 && dist > c.Rest*c.Tear {
		c.broken = true
		return
	}

	// Fractional error along the connecting axis.
	diff := (dist - c.Rest) / dist * c.Stiffness
	corr := delta.Scale(diff)

	// Distribute the correction by inverse mass.
	c.A.Pos = c.A.Pos.Add(corr.Scale(c.A.InvMass / invSum))
	c.B.Pos = c.B.Pos.Sub(corr.Scale(c.B.InvMass / invSum))
}
