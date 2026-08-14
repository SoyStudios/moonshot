package physics

import (
	"math"
	"testing"

	"github.com/SoyStudios/moonshot/crochet/math3"
)

// A free particle under gravity should fall (Y decreases).
func TestGravityPullsParticleDown(t *testing.T) {
	w := NewWorld()
	i := w.Add(math3.V(0, 10, 0), 1)
	for n := 0; n < 60; n++ {
		w.Step(1.0 / 60)
	}
	if w.Particles[i].Pos.Y >= 10 {
		t.Fatalf("particle did not fall: y=%v", w.Particles[i].Pos.Y)
	}
}

// The default rest threshold must not cancel gravity: a free particle at the
// demo's dt still falls a meaningful distance (guards the sleep-vs-gravity bug).
func TestRestThresholdDoesNotBreakFreeFall(t *testing.T) {
	w := NewWorld() // default Damping/RestThreshold
	i := w.Add(math3.V(0, 100, 0), 1)
	for n := 0; n < 120; n++ { // 1 second at dt=1/120
		w.Step(1.0 / 120)
	}
	if drop := 100 - w.Particles[i].Pos.Y; drop < 1.0 {
		t.Fatalf("particle barely fell (drop=%.3f); rest threshold likely cancelling gravity", drop)
	}
}

// A hanging strand must actually come to rest, not quiver forever.
func TestHangingStrandComesToRest(t *testing.T) {
	w := NewWorld()
	top := w.Add(math3.V(0, 10, 0), 1)
	w.Particles[top].Pin()
	prev := top
	for k := 0; k < 12; k++ {
		n := w.Add(math3.V(0, 10-0.3*float64(k+1), 0), 1)
		w.Link(prev, n, 1)
		prev = n
	}
	for n := 0; n < 1200; n++ {
		w.Step(1.0 / 120)
	}
	var speed float64
	for _, p := range w.Particles {
		speed += p.Velocity().Len()
	}
	if speed > 1e-4 {
		t.Fatalf("strand did not settle: total speed %.6f", speed)
	}
}

// A pinned particle must never move.
func TestPinnedStaysPut(t *testing.T) {
	w := NewWorld()
	i := w.Add(math3.V(1, 5, 2), 1)
	w.Particles[i].Pin()
	start := w.Particles[i].Pos
	for n := 0; n < 100; n++ {
		w.Step(1.0 / 60)
	}
	if w.Particles[i].Pos != start {
		t.Fatalf("pinned particle moved to %+v", w.Particles[i].Pos)
	}
}

// A two-particle strand hung from a pin should settle close to the rest length
// below the pin — the distance constraint must resist gravity's stretch.
func TestDistanceConstraintHoldsRestLength(t *testing.T) {
	w := NewWorld()
	w.Iterations = 40
	top := w.Add(math3.V(0, 10, 0), 1)
	bot := w.Add(math3.V(0, 9, 0), 1) // rest length 1
	w.Particles[top].Pin()
	c := w.Link(top, bot, 1)

	for n := 0; n < 600; n++ {
		w.Step(1.0 / 60)
	}

	got := w.Particles[top].Pos.Distance(w.Particles[bot].Pos)
	if math.Abs(got-c.Rest) > 0.02 {
		t.Fatalf("settled length = %v, want ~%v (rest)", got, c.Rest)
	}
	// It should hang essentially straight down.
	if w.Particles[bot].Pos.Y > w.Particles[top].Pos.Y {
		t.Fatalf("bottom particle rose above the pin")
	}
}

// The ground plane must stop a falling particle.
func TestGroundStopsFall(t *testing.T) {
	w := NewWorld()
	w.Ground = true
	w.GroundY = 0
	i := w.Add(math3.V(0, 5, 0), 1)
	for n := 0; n < 300; n++ {
		w.Step(1.0 / 60)
	}
	if w.Particles[i].Pos.Y < -1e-6 {
		t.Fatalf("particle fell through ground: y=%v", w.Particles[i].Pos.Y)
	}
}

// A tearable constraint under a hard pull should snap.
func TestTearBreaksConstraint(t *testing.T) {
	w := NewWorld()
	a := w.Add(math3.V(0, 0, 0), 1)
	b := w.Add(math3.V(0, 1, 0), 1)
	w.Particles[a].Pin()
	c := w.Link(a, b, 1)
	c.Tear = 1.5
	// Yank b far past the tear threshold, then step.
	w.Particles[b].Pos = math3.V(0, 5, 0)
	w.Step(1.0 / 60)
	if !c.Broken() {
		t.Fatalf("constraint should have torn")
	}
}
