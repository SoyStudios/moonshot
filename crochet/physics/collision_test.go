package physics

import (
	"math"
	"testing"

	"github.com/SoyStudios/moonshot/crochet/math3"
)

func TestSelfCollisionPushesApart(t *testing.T) {
	w := NewWorld()
	w.Gravity = math3.Zero
	w.SelfCollision = true
	w.CollisionRadius = 0.5 // min separation 1.0
	a := w.Add(math3.V(0, 0, 0), 1)
	b := w.Add(math3.V(0.2, 0, 0), 1) // overlapping

	for n := 0; n < 50; n++ {
		w.Step(1.0 / 60)
	}

	got := w.Particles[a].Pos.Distance(w.Particles[b].Pos)
	if got < 2*w.CollisionRadius-1e-3 {
		t.Fatalf("particles still overlapping: dist=%v want >= %v", got, 2*w.CollisionRadius)
	}
}

func TestSelfCollisionSkipsLinkedPair(t *testing.T) {
	w := NewWorld()
	w.Gravity = math3.Zero
	w.SelfCollision = true
	w.CollisionRadius = 0.5 // min separation 1.0, larger than the rest length
	a := w.Add(math3.V(0, 0, 0), 1)
	b := w.Add(math3.V(0.3, 0, 0), 1)
	w.Particles[a].Pin()
	c := w.Link(a, b, 1) // rest length 0.3

	for n := 0; n < 100; n++ {
		w.Step(1.0 / 60)
	}

	got := w.Particles[a].Pos.Distance(w.Particles[b].Pos)
	if got > c.Rest+0.05 {
		t.Fatalf("linked pair was pushed apart by collision: dist=%v want ~%v", got, c.Rest)
	}
}

func TestPressureInflatesRing(t *testing.T) {
	w := NewWorld()
	w.Gravity = math3.Zero
	// A small ring of particles linked into a loop.
	n := 8
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		a := float64(i) / float64(n) * 2 * math.Pi
		idx[i] = w.Add(math3.V(math.Cos(a), 0, math.Sin(a)), 1)
	}
	members := make([]*Particle, n)
	var before float64
	for i := 0; i < n; i++ {
		members[i] = w.Particles[idx[i]]
		w.Link(idx[i], idx[(i+1)%n], 1)
		before += members[i].Pos.Len()
	}
	w.AddConstraint(NewPressure(members, 0.01))

	for k := 0; k < 200; k++ {
		w.Step(1.0 / 60)
	}
	var after float64
	for _, m := range members {
		after += m.Pos.Len()
	}
	if after <= before {
		t.Fatalf("pressure did not inflate the ring: before=%v after=%v", before, after)
	}
}
