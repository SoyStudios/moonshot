package physics

import "github.com/SoyStudios/moonshot/crochet/math3"

// Sphere is a static spherical collider that particles are pushed out of. It is
// handy for draping crochet fabric over a form (a ball, a head, a stuffed toy).
type Sphere struct {
	Center math3.Vec3
	Radius float64
}

// World owns the particles and constraints and advances the simulation. It uses
// Verlet integration followed by several constraint-relaxation iterations
// (Position Based Dynamics), which is stable, cheap and well suited to the
// stretch-resistant, inextensible feel of yarn.
type World struct {
	Particles   []*Particle
	Constraints []Constraint

	Gravity math3.Vec3 // acceleration applied to every free particle
	Damping float64    // velocity retention per step in (0,1]; <1 bleeds energy

	// Iterations is the number of constraint relaxation passes per step. Higher
	// values make yarn less stretchy at the cost of CPU.
	Iterations int

	// Ground, when enabled, is an infinite horizontal plane at GroundY that
	// particles cannot fall through. Friction in [0,1] damps sliding on it.
	Ground   bool
	GroundY  float64
	Friction float64

	Spheres []Sphere
}

// NewWorld returns a world with sensible defaults for yarn simulation.
func NewWorld() *World {
	return &World{
		Gravity:    math3.V(0, -9.81, 0),
		Damping:    0.99,
		Iterations: 12,
		GroundY:    0,
		Friction:   0.35,
	}
}

// AddParticle appends a particle and returns its index. Indices are stable and
// used by the yarn/pattern layers to refer to nodes.
func (w *World) AddParticle(p *Particle) int {
	w.Particles = append(w.Particles, p)
	return len(w.Particles) - 1
}

// Add is a convenience that creates and appends a particle, returning its index.
func (w *World) Add(pos math3.Vec3, mass float64) int {
	return w.AddParticle(NewParticle(pos, mass))
}

// AddConstraint appends a constraint to be solved each step.
func (w *World) AddConstraint(c Constraint) { w.Constraints = append(w.Constraints, c) }

// Link creates and registers a distance constraint between two particle indices
// at their current separation, returning it so callers can tune Tear/Rest.
func (w *World) Link(a, b int, stiffness float64) *DistanceConstraint {
	c := NewDistanceConstraint(w.Particles[a], w.Particles[b], stiffness)
	w.AddConstraint(c)
	return c
}

// AddForceAll applies a force to every free particle (e.g. a wind gust).
func (w *World) AddForceAll(f math3.Vec3) {
	for _, p := range w.Particles {
		if p.InvMass != 0 {
			p.AddForce(f)
		}
	}
}

// Step advances the simulation by dt seconds: integrate, then relax
// constraints and resolve collisions repeatedly.
func (w *World) Step(dt float64) {
	w.integrate(dt)
	for i := 0; i < w.Iterations; i++ {
		for _, c := range w.Constraints {
			c.Solve()
		}
		w.collide()
	}
}

func (w *World) integrate(dt float64) {
	dt2 := dt * dt
	for _, p := range w.Particles {
		if p.InvMass == 0 {
			p.force = math3.Zero
			continue
		}
		// Verlet: next = pos + (pos-prev)*damping + accel*dt^2
		vel := p.Pos.Sub(p.Prev).Scale(w.Damping)
		accel := w.Gravity.Add(p.force.Scale(p.InvMass))
		next := p.Pos.Add(vel).Add(accel.Scale(dt2))
		p.Prev = p.Pos
		p.Pos = next
		p.force = math3.Zero
	}
}

func (w *World) collide() {
	for _, p := range w.Particles {
		if p.InvMass == 0 {
			continue
		}
		// Sphere colliders: project the point onto the surface.
		for _, s := range w.Spheres {
			d := p.Pos.Sub(s.Center)
			dist := d.Len()
			if dist < s.Radius && dist > 0 {
				p.Pos = s.Center.Add(d.Scale(s.Radius / dist))
			}
		}
		// Ground plane with simple Coulomb-ish friction on the tangent.
		if w.Ground && p.Pos.Y < w.GroundY {
			p.Pos.Y = w.GroundY
			// Damp horizontal sliding by pulling Prev toward Pos.
			vx := p.Pos.X - p.Prev.X
			vz := p.Pos.Z - p.Prev.Z
			p.Prev.X = p.Pos.X - vx*(1-w.Friction)
			p.Prev.Z = p.Pos.Z - vz*(1-w.Friction)
		}
	}
}
