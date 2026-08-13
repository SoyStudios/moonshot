// Package pattern turns high-level crochet descriptions into physics geometry.
//
// A crochet fabric is modelled as a grid of stitch nodes (particles) bonded by
// yarn. Two kinds of bond are created:
//
//   - the continuous yarn *path* threaded through the stitches of each
//     row/round (rendered as a thick tube), and
//   - structural cross-links between a stitch and the stitch it was worked into
//     in the previous row/round (rendered thin) — this is what turns a 1-D
//     strand into 2-D fabric that drapes.
//
// The builders here produce the three shapes most crochet starts from: a flat
// swatch, a worked-in-the-round tube, and a flat increasing disc (the
// amigurumi base). All of them hand back a Fabric referencing the shared
// physics world, ready to simulate and render.
package pattern

import (
	"math"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// Fabric is a materialised crochet piece: the yarn paths and cross-links that
// live inside World, plus the nodes that were pinned.
type Fabric struct {
	World   *physics.World
	Strands []*yarn.Strand // thick yarn paths (rows / rounds)
	Links   [][2]int       // thin structural cross-links between rows
	Pins    []int          // pinned particle indices
	Radius  float64        // yarn radius (used to render Links)
	Color   yarn.Color
}

// PinMode selects which nodes of a swatch are pinned in place.
type PinMode int

const (
	PinNone       PinMode = iota // let it fall freely
	PinTopEdge                   // pin the whole first row (a hanging tapestry)
	PinTopCorners                // pin only the two ends of the first row
	PinCorners                   // pin all four corners (a stretched hammock)
)

// SwatchConfig describes a flat rectangular piece of crochet.
type SwatchConfig struct {
	Rows, Cols int        // stitch counts
	Origin     math3.Vec3 // position of stitch (0,0)
	U, V       math3.Vec3 // per-column and per-row step directions (world units)
	Stitch     Stitch     // stitch type (sets the effective row height via V)
	Pin        PinMode
	Yarn       yarn.Config
}

// Swatch builds a Rows×Cols grid. The yarn snakes back and forth across each
// row (boustrophedon, like real crochet turning at the end of a row); rows are
// bonded to the row below by vertical cross-links.
func Swatch(w *physics.World, cfg SwatchConfig) *Fabric {
	if cfg.Rows < 1 {
		cfg.Rows = 1
	}
	if cfg.Cols < 1 {
		cfg.Cols = 1
	}
	f := &Fabric{World: w, Radius: cfg.Yarn.Radius, Color: cfg.Yarn.Color}

	// Lay out the grid of particles.
	grid := make([][]int, cfg.Rows)
	for r := 0; r < cfg.Rows; r++ {
		grid[r] = make([]int, cfg.Cols)
		for c := 0; c < cfg.Cols; c++ {
			pos := cfg.Origin.
				Add(cfg.U.Scale(float64(c))).
				Add(cfg.V.Scale(float64(r)))
			grid[r][c] = w.Add(pos, cfg.Yarn.SegmentMass)
		}
	}

	// Thread one continuous yarn path through every stitch, boustrophedon.
	path := yarn.New(cfg.Yarn)
	for r := 0; r < cfg.Rows; r++ {
		if r%2 == 0 {
			for c := 0; c < cfg.Cols; c++ {
				path.Append(grid[r][c])
			}
		} else {
			for c := cfg.Cols - 1; c >= 0; c-- {
				path.Append(grid[r][c])
			}
		}
	}
	path.Connect(w, cfg.Yarn) // structural + bending along the path
	f.Strands = append(f.Strands, path)

	// Bond each stitch to the one below it (the "work into the previous row").
	for r := 1; r < cfg.Rows; r++ {
		for c := 0; c < cfg.Cols; c++ {
			w.Link(grid[r-1][c], grid[r][c], cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{grid[r-1][c], grid[r][c]})
		}
	}

	f.applyPins(pinsFor(cfg.Pin, grid))
	return f
}

// TubeConfig describes crochet worked in the round into a cylinder (a hat body,
// an amigurumi limb).
type TubeConfig struct {
	Rounds      int        // number of stacked rounds
	Stitches    int        // stitches per round
	Radius      float64    // tube radius
	RiseY       float64    // vertical rise per round
	Center      math3.Vec3 // centre of the bottom round
	PinTopRound bool       // pin the last round (hang the tube)
	Yarn        yarn.Config
}

// Tube builds a cylinder of stitches. Each round is a closed yarn loop; rounds
// are bonded vertically. A gentle per-round twist mimics the spiral of real
// worked-in-the-round crochet.
func Tube(w *physics.World, cfg TubeConfig) *Fabric {
	if cfg.Rounds < 1 {
		cfg.Rounds = 1
	}
	if cfg.Stitches < 3 {
		cfg.Stitches = 3
	}
	f := &Fabric{World: w, Radius: cfg.Yarn.Radius, Color: cfg.Yarn.Color}

	grid := make([][]int, cfg.Rounds)
	twist := (2 * math.Pi / float64(cfg.Stitches)) // one stitch of spiral per round
	for r := 0; r < cfg.Rounds; r++ {
		grid[r] = make([]int, cfg.Stitches)
		for i := 0; i < cfg.Stitches; i++ {
			a := 2*math.Pi*float64(i)/float64(cfg.Stitches) + twist*float64(r)
			pos := cfg.Center.Add(math3.V(
				cfg.Radius*math.Cos(a),
				cfg.RiseY*float64(r),
				cfg.Radius*math.Sin(a),
			))
			grid[r][i] = w.Add(pos, cfg.Yarn.SegmentMass)
		}
	}

	// Each round is a closed loop of yarn.
	for r := 0; r < cfg.Rounds; r++ {
		ring := yarn.New(cfg.Yarn)
		for i := 0; i < cfg.Stitches; i++ {
			ring.Append(grid[r][i])
		}
		ring.Append(grid[r][0]) // close the loop
		ring.Connect(w, cfg.Yarn)
		f.Strands = append(f.Strands, ring)
	}

	// Bond rounds vertically.
	for r := 1; r < cfg.Rounds; r++ {
		for i := 0; i < cfg.Stitches; i++ {
			w.Link(grid[r-1][i], grid[r][i], cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{grid[r-1][i], grid[r][i]})
		}
	}

	if cfg.PinTopRound {
		f.applyPins(grid[cfg.Rounds-1])
	}
	return f
}

// DiscConfig describes a flat disc worked in the round with increases each
// round — the classic amigurumi / granny-circle base.
type DiscConfig struct {
	Rounds      int        // number of rounds (excluding the centre)
	StartStitch int        // stitches in the first round (e.g. 6)
	Increase    int        // stitches added each subsequent round (e.g. 6)
	RingSpacing float64    // radial distance between rounds
	Center      math3.Vec3 // disc centre (in the XZ plane by default)
	PinCenter   bool       // pin the middle so the disc can drape from it
	Yarn        yarn.Config
}

// Disc builds a flat increasing circle lying in the XZ plane. Each round is a
// ring bonded radially to the round inside it.
func Disc(w *physics.World, cfg DiscConfig) *Fabric {
	if cfg.Rounds < 1 {
		cfg.Rounds = 1
	}
	if cfg.StartStitch < 3 {
		cfg.StartStitch = 3
	}
	f := &Fabric{World: w, Radius: cfg.Yarn.Radius, Color: cfg.Yarn.Color}

	center := w.Add(cfg.Center, cfg.Yarn.SegmentMass)
	rings := make([][]int, cfg.Rounds)
	for r := 0; r < cfg.Rounds; r++ {
		n := cfg.StartStitch + cfg.Increase*r
		radius := cfg.RingSpacing * float64(r+1)
		rings[r] = make([]int, n)
		for i := 0; i < n; i++ {
			a := 2 * math.Pi * float64(i) / float64(n)
			pos := cfg.Center.Add(math3.V(radius*math.Cos(a), 0, radius*math.Sin(a)))
			rings[r][i] = w.Add(pos, cfg.Yarn.SegmentMass)
		}

		// The ring itself is a closed yarn loop.
		ring := yarn.New(cfg.Yarn)
		for i := 0; i < n; i++ {
			ring.Append(rings[r][i])
		}
		ring.Append(rings[r][0])
		ring.Connect(w, cfg.Yarn)
		f.Strands = append(f.Strands, ring)
	}

	// Radial bonds: every stitch links to the nearest stitch one ring inward
	// (or the centre for the first ring).
	for r := 0; r < cfg.Rounds; r++ {
		n := len(rings[r])
		for i := 0; i < n; i++ {
			var inner int
			if r == 0 {
				inner = center
			} else {
				m := len(rings[r-1])
				j := int(math.Round(float64(i) / float64(n) * float64(m)))
				inner = rings[r-1][j%m]
			}
			w.Link(inner, rings[r][i], cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{inner, rings[r][i]})
		}
	}

	if cfg.PinCenter {
		f.applyPins([]int{center})
	}
	return f
}

// applyPins pins the given particle indices and records them.
func (f *Fabric) applyPins(idx []int) {
	for _, i := range idx {
		f.World.Particles[i].Pin()
		f.Pins = append(f.Pins, i)
	}
}

func pinsFor(mode PinMode, grid [][]int) []int {
	rows := len(grid)
	if rows == 0 {
		return nil
	}
	cols := len(grid[0])
	switch mode {
	case PinTopEdge:
		return append([]int(nil), grid[0]...)
	case PinTopCorners:
		return []int{grid[0][0], grid[0][cols-1]}
	case PinCorners:
		return []int{
			grid[0][0], grid[0][cols-1],
			grid[rows-1][0], grid[rows-1][cols-1],
		}
	default:
		return nil
	}
}
