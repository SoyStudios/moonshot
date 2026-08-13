package pattern

import (
	"math"
	"testing"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

func finite(v math3.Vec3) bool {
	return !math.IsNaN(v.X) && !math.IsNaN(v.Y) && !math.IsNaN(v.Z) &&
		!math.IsInf(v.X, 0) && !math.IsInf(v.Y, 0) && !math.IsInf(v.Z, 0)
}

func simulate(w *physics.World, steps int) {
	for i := 0; i < steps; i++ {
		w.Step(1.0 / 120)
	}
}

func TestSwatchBuildsGridAndPins(t *testing.T) {
	w := physics.NewWorld()
	f := Swatch(w, SwatchConfig{
		Rows: 8, Cols: 10,
		Origin: math3.V(0, 5, 0),
		U:      math3.V(0.5, 0, 0),
		V:      math3.V(0, -0.5, 0),
		Pin:    PinTopEdge,
		Yarn:   yarn.DefaultConfig(),
	})

	if got := len(w.Particles); got != 8*10 {
		t.Fatalf("expected 80 particles, got %d", got)
	}
	if len(f.Pins) != 10 {
		t.Fatalf("PinTopEdge should pin 10 nodes, got %d", len(f.Pins))
	}
	// Pinned nodes must be immovable after simulation.
	before := make([]math3.Vec3, len(f.Pins))
	for i, p := range f.Pins {
		before[i] = w.Particles[p].Pos
	}
	simulate(w, 300)
	for i, p := range f.Pins {
		if w.Particles[p].Pos != before[i] {
			t.Fatalf("pinned node %d moved", p)
		}
	}
}

func TestFabricStaysFinite(t *testing.T) {
	cases := map[string]*physics.World{}

	w1 := physics.NewWorld()
	w1.Ground = true
	Swatch(w1, SwatchConfig{Rows: 12, Cols: 12, Origin: math3.V(-3, 8, 0),
		U: math3.V(0.5, 0, 0), V: math3.V(0, -0.5, 0), Pin: PinTopCorners, Yarn: yarn.DefaultConfig()})
	cases["swatch"] = w1

	w2 := physics.NewWorld()
	w2.Ground = true
	Tube(w2, TubeConfig{Rounds: 12, Stitches: 20, Radius: 2, RiseY: 0.4,
		Center: math3.V(0, 1, 0), PinTopRound: true, Yarn: yarn.DefaultConfig()})
	cases["tube"] = w2

	w3 := physics.NewWorld()
	w3.Ground = true
	Disc(w3, DiscConfig{Rounds: 6, StartStitch: 6, Increase: 6, RingSpacing: 0.6,
		Center: math3.V(0, 5, 0), PinCenter: true, Yarn: yarn.DefaultConfig()})
	cases["disc"] = w3

	for name, w := range cases {
		simulate(w, 600)
		for i, p := range w.Particles {
			if !finite(p.Pos) {
				t.Fatalf("%s: particle %d went non-finite: %+v", name, i, p.Pos)
			}
		}
	}
}

// A hanging swatch should settle: the average speed of its free nodes drops
// close to zero once the yarn stops swinging.
func TestSwatchSettles(t *testing.T) {
	w := physics.NewWorld()
	f := Swatch(w, SwatchConfig{
		Rows: 10, Cols: 12,
		Origin: math3.V(-3, 8, 0),
		U:      math3.V(0.5, 0, 0),
		V:      math3.V(0, -0.5, 0),
		Pin:    PinTopEdge,
		Yarn:   yarn.DefaultConfig(),
	})
	simulate(w, 1500)

	var speed float64
	var free int
	for _, p := range w.Particles {
		if p.Pinned() {
			continue
		}
		speed += p.Velocity().Len()
		free++
	}
	avg := speed / float64(free)
	if avg > 0.02 {
		t.Fatalf("swatch did not settle: avg step speed %.4f", avg)
	}
	_ = f
}
