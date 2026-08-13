package physics

import "github.com/SoyStudios/moonshot/crochet/math3"

// Particle is a point mass integrated with Verlet integration. Rather than
// storing an explicit velocity, velocity is implied by the difference between
// the current position (Pos) and the previous position (Prev). This makes the
// integrator stable and makes position-based constraints trivial to apply.
type Particle struct {
	Pos  math3.Vec3 // current position
	Prev math3.Vec3 // position on the previous step (encodes velocity)

	// InvMass is 1/mass. A value of 0 means infinite mass: the particle is
	// pinned and never moved by integration, constraints or collisions.
	InvMass float64

	// force accumulates external forces (e.g. wind) applied this step. It is
	// cleared after every integration step.
	force math3.Vec3
}

// NewParticle creates a particle at pos with the given mass. A mass <= 0 pins
// the particle in place.
func NewParticle(pos math3.Vec3, mass float64) *Particle {
	p := &Particle{Pos: pos, Prev: pos}
	if mass > 0 {
		p.InvMass = 1 / mass
	}
	return p
}

// Pinned reports whether the particle is immovable.
func (p *Particle) Pinned() bool { return p.InvMass == 0 }

// Pin fixes the particle in place at its current position.
func (p *Particle) Pin() { p.InvMass = 0 }

// SetMass changes the particle mass. A mass <= 0 pins it.
func (p *Particle) SetMass(mass float64) {
	if mass > 0 {
		p.InvMass = 1 / mass
	} else {
		p.InvMass = 0
	}
}

// AddForce accumulates an external force to be applied on the next step.
func (p *Particle) AddForce(f math3.Vec3) { p.force = p.force.Add(f) }

// Velocity returns the implicit velocity per step (Pos - Prev). It is not a
// true per-second velocity; multiply by 1/dt if you need that.
func (p *Particle) Velocity() math3.Vec3 { return p.Pos.Sub(p.Prev) }
