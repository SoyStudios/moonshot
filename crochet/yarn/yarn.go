// Package yarn models a physical strand of yarn on top of the physics solver.
//
// A Strand is an ordered list of particle indices (its "nodes") that share a
// radius and colour. Consecutive nodes are joined by distance constraints (the
// yarn's inextensibility) and, optionally, weaker "skip-one" constraints that
// resist sharp folding (bending stiffness). This is the primitive the crochet
// pattern layer stitches together into fabric.
package yarn

import (
	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/physics"
)

// Color is a simple RGBA colour kept free of any rendering dependency so the
// yarn/pattern layers stay headless-testable. The engine maps it to raylib.
type Color struct{ R, G, B, A uint8 }

// Config controls how a strand is materialised into the physics world.
type Config struct {
	SegmentMass float64 // mass per node
	Stiffness   float64 // structural stiffness in [0,1]
	Bending     float64 // skip-one stiffness in [0,1]; 0 disables bending
	Radius      float64 // render thickness
	Color       Color
}

// DefaultConfig returns reasonable yarn parameters.
func DefaultConfig() Config {
	return Config{
		SegmentMass: 0.05,
		Stiffness:   1.0,
		Bending:     0.15,
		Radius:      0.06,
		Color:       Color{220, 120, 80, 255},
	}
}

// Strand is a continuous piece of yarn threaded through world particles.
type Strand struct {
	Nodes  []int
	Radius float64
	Color  Color
}

// New creates an (initially empty) strand with the given appearance.
func New(cfg Config) *Strand {
	return &Strand{Radius: cfg.Radius, Color: cfg.Color}
}

// Line samples a straight run of yarn between a and b into segments+1 nodes,
// wiring up the structural and bending constraints, and returns the strand.
// The first and/or last node can be pinned to hang the yarn.
func Line(w *physics.World, a, b math3.Vec3, segments int, cfg Config) *Strand {
	s := New(cfg)
	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		idx := w.Add(math3.Lerp(a, b, t), cfg.SegmentMass)
		s.Nodes = append(s.Nodes, idx)
	}
	s.Connect(w, cfg)
	return s
}

// Append adds an already-created particle index as the next node of the strand.
func (s *Strand) Append(idx int) { s.Nodes = append(s.Nodes, idx) }

// Connect wires the strand's current node list with structural (consecutive)
// and optional bending (skip-one) distance constraints. Call it after all nodes
// have been appended.
func (s *Strand) Connect(w *physics.World, cfg Config) {
	for i := 0; i+1 < len(s.Nodes); i++ {
		w.Link(s.Nodes[i], s.Nodes[i+1], cfg.Stiffness)
	}
	if cfg.Bending > 0 {
		for i := 0; i+2 < len(s.Nodes); i++ {
			w.Link(s.Nodes[i], s.Nodes[i+2], cfg.Bending)
		}
	}
}

// Segments returns the ordered pairs of particle indices that make up the
// visible yarn path, for rendering as a sequence of tube sections.
func (s *Strand) Segments() [][2]int {
	out := make([][2]int, 0, len(s.Nodes))
	for i := 0; i+1 < len(s.Nodes); i++ {
		out = append(out, [2]int{s.Nodes[i], s.Nodes[i+1]})
	}
	return out
}
