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
// The builders here produce the shapes most crochet starts from: a flat swatch,
// a worked-in-the-round tube, a flat increasing disc (the amigurumi base), and
// a general surface of revolution driven by per-round stitch counts (spheres,
// cones, teardrops — the amigurumi vocabulary). All of them hand back a Fabric
// referencing the shared physics world, ready to simulate and render.
package pattern

import (
	"math"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// StitchCell describes one stitch for rendering it as a recognisable crochet
// "V" rather than a bare lattice node. It names the stitch's particle plus its
// neighbours in the fabric, from which the renderer derives the local surface
// frame (row direction × column direction → outward normal) and draws a small
// bowed V of the right size and colour. It is purely cosmetic — the physics
// still sees one particle per stitch.
type StitchCell struct {
	Node                  int        // this stitch's particle
	Left, Right, Down, Up int        // neighbour particles, or -1 if absent
	W, H                  float64    // stitch width and height in world units
	Color                 yarn.Color // banded/striped colour of this stitch
}

// Fabric is a materialised crochet piece: the yarn paths and cross-links that
// live inside World, plus the nodes that were pinned.
type Fabric struct {
	World    *physics.World
	Strands  []*yarn.Strand // thick yarn paths (rows / rounds)
	Links    [][2]int       // thin structural cross-links between rows
	Cells    []StitchCell   // per-stitch cells for crochet-style rendering
	Pins     []int          // pinned particle indices
	Nodes    []int          // every particle this fabric created
	Radius   float64        // yarn radius (used to render Links)
	Color    yarn.Color
	Material yarn.Material

	// Radial / Cylindrical mark closed shapes worked in the round so the
	// renderer can orient each stitch from the smooth outward surface normal
	// instead of a noisy neighbour cross product. Radial (a ball) uses the full
	// point−centroid direction; Cylindrical (a tube) removes the vertical axis
	// so the normal points straight out from the side. Flat pieces leave both
	// false and use the neighbour frame.
	Radial      bool
	Cylindrical bool
}

// newFabric starts a fabric from a yarn config.
func newFabric(w *physics.World, cfg yarn.Config) *Fabric {
	mat := cfg.Material
	if mat == (yarn.Material{}) {
		mat = yarn.DefaultMaterial()
	}
	return &Fabric{World: w, Radius: cfg.Radius, Color: cfg.Color, Material: mat}
}

// add creates a particle, records it as part of the fabric, and returns its index.
func (f *Fabric) add(pos math3.Vec3, mass float64) int {
	i := f.World.Add(pos, mass)
	f.Nodes = append(f.Nodes, i)
	return i
}

// Particles returns the physics particles this fabric owns — the target for a
// stuffing constraint, mass tweaks, etc.
func (f *Fabric) Particles() []*physics.Particle {
	out := make([]*physics.Particle, len(f.Nodes))
	for i, n := range f.Nodes {
		out[i] = f.World.Particles[n]
	}
	return out
}

// Stuff adds a stuffing (internal pressure) constraint over the whole fabric
// and returns it, so a closed shape inflates like a stuffed amigurumi.
func (f *Fabric) Stuff(strength float64) *physics.PressureConstraint {
	c := physics.NewPressure(f.Particles(), strength)
	f.World.AddConstraint(c)
	return c
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
	Stitch     Stitch     // stitch type (documented gauge; see Def)
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
	f := newFabric(w, cfg.Yarn)

	grid := make([][]int, cfg.Rows)
	for r := 0; r < cfg.Rows; r++ {
		grid[r] = make([]int, cfg.Cols)
		for c := 0; c < cfg.Cols; c++ {
			pos := cfg.Origin.
				Add(cfg.U.Scale(float64(c))).
				Add(cfg.V.Scale(float64(r)))
			grid[r][c] = f.add(pos, cfg.Yarn.SegmentMass)
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
	path.Connect(w, cfg.Yarn)
	f.Strands = append(f.Strands, path)

	// Bond each stitch to the one below it (the "work into the previous row").
	for r := 1; r < cfg.Rows; r++ {
		for c := 0; c < cfg.Cols; c++ {
			w.Link(grid[r-1][c], grid[r][c], cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{grid[r-1][c], grid[r][c]})
		}
	}

	f.applyPins(pinsFor(cfg.Pin, grid))
	f.buildGridCells(grid, cfg.U.Len(), cfg.V.Len(), cfg.Yarn)
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
	f := newFabric(w, cfg.Yarn)
	f.Cylindrical = true

	grid := make([][]int, cfg.Rounds)
	twist := 2 * math.Pi / float64(cfg.Stitches)
	for r := 0; r < cfg.Rounds; r++ {
		grid[r] = make([]int, cfg.Stitches)
		for i := 0; i < cfg.Stitches; i++ {
			a := 2*math.Pi*float64(i)/float64(cfg.Stitches) + twist*float64(r)
			pos := cfg.Center.Add(math3.V(
				cfg.Radius*math.Cos(a),
				cfg.RiseY*float64(r),
				cfg.Radius*math.Sin(a),
			))
			grid[r][i] = f.add(pos, cfg.Yarn.SegmentMass)
		}
	}

	for r := 0; r < cfg.Rounds; r++ {
		ring := ringStrand(w, grid[r], cfg.Yarn)
		bandRound(ring, cfg.Yarn, r)
		f.Strands = append(f.Strands, ring)
	}
	for r := 1; r < cfg.Rounds; r++ {
		for i := 0; i < cfg.Stitches; i++ {
			w.Link(grid[r-1][i], grid[r][i], cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{grid[r-1][i], grid[r][i]})
		}
	}

	if cfg.PinTopRound {
		f.applyPins(grid[cfg.Rounds-1])
	}
	radii := make([]float64, cfg.Rounds)
	for r := range radii {
		radii[r] = cfg.Radius
	}
	f.buildRingCells(grid, radii, cfg.RiseY, cfg.Yarn)
	return f
}

// DiscConfig describes a flat disc worked in the round with increases each
// round — the classic amigurumi / granny-circle base.
type DiscConfig struct {
	Rounds      int
	StartStitch int
	Increase    int
	RingSpacing float64
	Center      math3.Vec3
	PinCenter   bool
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
	f := newFabric(w, cfg.Yarn)

	center := f.add(cfg.Center, cfg.Yarn.SegmentMass)
	rings := make([][]int, cfg.Rounds)
	radii := make([]float64, cfg.Rounds)
	for r := 0; r < cfg.Rounds; r++ {
		n := cfg.StartStitch + cfg.Increase*r
		radius := cfg.RingSpacing * float64(r+1)
		radii[r] = radius
		rings[r] = make([]int, n)
		for i := 0; i < n; i++ {
			a := 2 * math.Pi * float64(i) / float64(n)
			pos := cfg.Center.Add(math3.V(radius*math.Cos(a), 0, radius*math.Sin(a)))
			rings[r][i] = f.add(pos, cfg.Yarn.SegmentMass)
		}
		ring := ringStrand(w, rings[r], cfg.Yarn)
		bandRound(ring, cfg.Yarn, r)
		f.Strands = append(f.Strands, ring)
	}

	for r := 0; r < cfg.Rounds; r++ {
		linkToInner(f, rings[r], innerRing(rings, r, center), cfg.Yarn.Stiffness)
	}

	if cfg.PinCenter {
		f.applyPins([]int{center})
	}
	f.buildRingCells(rings, radii, cfg.RingSpacing, cfg.Yarn)
	return f
}

// RevolveConfig describes a surface of revolution built from a sequence of
// per-round stitch counts — the way amigurumi is actually written ("6, 12, 18,
// … , 12, 6"). Each round's circumference follows its stitch count, so
// increases bulge the shape out and decreases pull it in; the radius/rise are
// derived so the slant between rounds matches the stitch height. Feed it a
// bell curve of counts to get a ball, a ramp to get a cone.
type RevolveConfig struct {
	Counts      []int      // stitches per round, bottom → top
	Stitch      Stitch     // sets the gauge (width & height)
	Gauge       float64    // world units per stitch unit
	Center      math3.Vec3 // position of the bottom pole
	CloseBottom bool       // cinch the first round to a pole node (magic ring)
	CloseTop    bool       // cinch the last round to a pole node
	PinTop      bool       // pin the top pole / last round (hang it)
	Yarn        yarn.Config
}

// Revolve builds the surface described by RevolveConfig. Rings are stacked
// along +Y; each stitch bonds to the nearest stitch in the ring below, and
// optional pole nodes close the ends so a stuffing constraint can pressurise a
// sealed shape.
func Revolve(w *physics.World, cfg RevolveConfig) *Fabric {
	f := newFabric(w, cfg.Yarn)
	f.Radial = true
	if len(cfg.Counts) == 0 {
		return f
	}
	if cfg.Gauge <= 0 {
		cfg.Gauge = 0.5
	}
	def := cfg.Stitch.Def()
	width := cfg.Gauge * def.Width
	rowH := cfg.Gauge * def.Height

	// Radius of each round from its circumference, and the stacked Y so the
	// slant distance between successive rings equals the row height.
	nr := len(cfg.Counts)
	radius := make([]float64, nr)
	y := make([]float64, nr)
	for r := 0; r < nr; r++ {
		n := cfg.Counts[r]
		if n < 1 {
			n = 1
		}
		radius[r] = float64(n) * width / (2 * math.Pi)
		if r > 0 {
			dr := radius[r] - radius[r-1]
			// Guarantee some vertical progress even on a big increase/decrease.
			dy := math.Sqrt(math.Max(rowH*rowH-dr*dr, 0.04*rowH*rowH))
			y[r] = y[r-1] + dy
		}
	}

	rings := make([][]int, nr)
	for r := 0; r < nr; r++ {
		n := cfg.Counts[r]
		rings[r] = make([]int, n)
		for i := 0; i < n; i++ {
			// No per-round spiral offset: keeping rounds phase-aligned lets the
			// rendered stitches form clean vertical columns on the surface.
			a := 2 * math.Pi * float64(i) / float64(n)
			pos := cfg.Center.Add(math3.V(
				radius[r]*math.Cos(a),
				y[r],
				radius[r]*math.Sin(a),
			))
			rings[r][i] = f.add(pos, cfg.Yarn.SegmentMass)
		}
		ring := ringStrand(w, rings[r], cfg.Yarn)
		bandRound(ring, cfg.Yarn, r)
		f.Strands = append(f.Strands, ring)
	}

	// Vertical bonds between successive rings (nearest stitch below).
	for r := 1; r < nr; r++ {
		linkToInner(f, rings[r], rings[r-1], cfg.Yarn.Stiffness)
	}

	var bottomPole, topPole = -1, -1
	if cfg.CloseBottom {
		bottomPole = f.add(cfg.Center.Add(math3.V(0, y[0]-rowH*0.5, 0)), cfg.Yarn.SegmentMass)
		for _, n := range rings[0] {
			w.Link(bottomPole, n, cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{bottomPole, n})
		}
	}
	if cfg.CloseTop {
		topPole = f.add(cfg.Center.Add(math3.V(0, y[nr-1]+rowH*0.5, 0)), cfg.Yarn.SegmentMass)
		for _, n := range rings[nr-1] {
			w.Link(topPole, n, cfg.Yarn.Stiffness)
			f.Links = append(f.Links, [2]int{topPole, n})
		}
	}

	if cfg.PinTop {
		if topPole >= 0 {
			f.applyPins([]int{topPole})
		} else {
			f.applyPins(rings[nr-1])
		}
	}
	f.buildRingCells(rings, radius, rowH, cfg.Yarn)
	return f
}

// SphereCounts builds the classic amigurumi round counts for a ball: increase
// by `step` each round up to `rounds` rounds of increases, hold, then mirror
// back down. The peak count is rounds*step.
func SphereCounts(rounds, step int) []int {
	if rounds < 1 {
		rounds = 1
	}
	if step < 1 {
		step = 1
	}
	up := make([]int, rounds)
	for r := 0; r < rounds; r++ {
		up[r] = step * (r + 1)
	}
	// up = step, 2step, ..., rounds*step ; mirror for the decreases.
	counts := append([]int{}, up...)
	for r := rounds - 2; r >= 0; r-- {
		counts = append(counts, up[r])
	}
	return counts
}

// --- shared helpers ---

// ringStrand creates a closed loop strand over the given node ring and wires
// its constraints.
func ringStrand(w *physics.World, ring []int, cfg yarn.Config) *yarn.Strand {
	s := yarn.New(cfg)
	for _, n := range ring {
		s.Append(n)
	}
	if len(ring) > 0 {
		s.Append(ring[0]) // close the loop
	}
	s.Connect(w, cfg)
	return s
}

// stripeColor returns the colour of band index (row or round), honouring a
// stripe palette; falls back to the base colour when no stripe is set.
func stripeColor(cfg yarn.Config, band int) yarn.Color {
	if len(cfg.Stripe) == 0 {
		return cfg.Color
	}
	w := cfg.StripeWidth
	if w < 1 {
		w = 1
	}
	return cfg.Stripe[(band/w)%len(cfg.Stripe)]
}

// buildGridCells emits one StitchCell per node of a rectangular grid, wiring up
// the four orthogonal neighbours (−1 where the grid ends).
func (f *Fabric) buildGridCells(grid [][]int, w, h float64, cfg yarn.Config) {
	rows := len(grid)
	if rows == 0 {
		return
	}
	cols := len(grid[0])
	at := func(r, c int) int {
		if r < 0 || r >= rows || c < 0 || c >= cols {
			return -1
		}
		return grid[r][c]
	}
	for r := 0; r < rows; r++ {
		col := stripeColor(cfg, r)
		for c := 0; c < cols; c++ {
			f.Cells = append(f.Cells, StitchCell{
				Node: grid[r][c],
				Left: at(r, c-1), Right: at(r, c+1),
				Down: at(r-1, c), Up: at(r+1, c),
				W: w, H: h, Color: col,
			})
		}
	}
}

// buildRingCells emits cells for a stack of rings. Left/Right wrap around each
// ring; Down/Up map to the nearest stitch in the adjacent ring.
func (f *Fabric) buildRingCells(rings [][]int, radius []float64, h float64, cfg yarn.Config) {
	nearest := func(i, n, m int) int { return int(math.Round(float64(i)/float64(n)*float64(m))) % m }
	for r := 0; r < len(rings); r++ {
		n := len(rings[r])
		if n == 0 {
			continue
		}
		w := 2 * math.Pi * radius[r] / float64(n)
		col := stripeColor(cfg, r)
		for i := 0; i < n; i++ {
			down, up := -1, -1
			if r > 0 && len(rings[r-1]) > 0 {
				down = rings[r-1][nearest(i, n, len(rings[r-1]))]
			}
			if r+1 < len(rings) && len(rings[r+1]) > 0 {
				up = rings[r+1][nearest(i, n, len(rings[r+1]))]
			}
			f.Cells = append(f.Cells, StitchCell{
				Node:  rings[r][i],
				Left:  rings[r][(i-1+n)%n],
				Right: rings[r][(i+1)%n],
				Down:  down, Up: up,
				W: w, H: h, Color: col,
			})
		}
	}
}

// bandRound recolours a round's strand to a solid stripe colour when the yarn
// config defines a stripe palette — giving worked-in-the-round pieces clean
// horizontal colour bands, StripeWidth rounds tall.
func bandRound(s *yarn.Strand, cfg yarn.Config, round int) {
	if len(cfg.Stripe) == 0 {
		return
	}
	band := cfg.StripeWidth
	if band < 1 {
		band = 1
	}
	s.Recolor(cfg.Stripe[(round/band)%len(cfg.Stripe)])
}

// innerRing returns the ring immediately inside r, or a single-element ring of
// the centre node for r == 0.
func innerRing(rings [][]int, r, center int) []int {
	if r == 0 {
		return []int{center}
	}
	return rings[r-1]
}

// linkToInner bonds every stitch of `outer` to the nearest stitch of `inner`.
func linkToInner(f *Fabric, outer, inner []int, stiffness float64) {
	n, m := len(outer), len(inner)
	if n == 0 || m == 0 {
		return
	}
	for i := 0; i < n; i++ {
		j := int(math.Round(float64(i)/float64(n)*float64(m))) % m
		f.World.Link(inner[j], outer[i], stiffness)
		f.Links = append(f.Links, [2]int{inner[j], outer[i]})
	}
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
