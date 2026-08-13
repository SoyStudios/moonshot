package pattern

import (
	"math"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// This file adds a small crochet "pattern" model: the actual language amigurumi
// is written in — rounds of operations worked into the stitches of the round
// below, with increases (two stitches in one) and decreases (one stitch across
// two). Build turns such a pattern into a Fabric, anchoring every stitch to the
// specific stitch(es) it is worked into, so the increase/decrease topology is
// correct rather than guessed. Parse (see parse.go) reads the same thing from
// standard written notation.

// Op is one crochet operation within a round.
type Op int

const (
	OpSc   Op = iota // single crochet: 1 stitch worked into 1 stitch below
	OpInc            // increase: 2 stitches into 1 below
	OpDec            // decrease: 1 stitch across 2 below
	OpHdc            // half double crochet (taller), 1 into 1
	OpDc             // double crochet (taller), 1 into 1
	OpTr             // treble crochet (tallest), 1 into 1
	OpSlSt           // slip stitch, 1 into 1
)

// spec returns how many new stitches the op makes, how many stitches below it
// consumes, and the base stitch (for row height).
func (o Op) spec() (newStitches, anchors int, st Stitch) {
	switch o {
	case OpInc:
		return 2, 1, Single
	case OpDec:
		return 1, 2, Single
	case OpHdc:
		return 1, 1, HalfDouble
	case OpDc:
		return 1, 1, Double
	case OpTr:
		return 1, 1, Treble
	case OpSlSt:
		return 1, 1, Slip
	default: // OpSc
		return 1, 1, Single
	}
}

func (o Op) String() string {
	switch o {
	case OpInc:
		return "inc"
	case OpDec:
		return "dec"
	case OpHdc:
		return "hdc"
	case OpDc:
		return "dc"
	case OpTr:
		return "tr"
	case OpSlSt:
		return "sl st"
	default:
		return "sc"
	}
}

func stitchOp(s Stitch) Op {
	switch s {
	case HalfDouble:
		return OpHdc
	case Double:
		return OpDc
	case Treble:
		return OpTr
	case Slip:
		return OpSlSt
	default:
		return OpSc
	}
}

// Round is one round of a pattern. Either an explicit list of Ops, or a Fill
// round that works one FillStitch into every stitch of the round below (the
// common "sc in each st around").
type Round struct {
	Ops        []Op
	Fill       bool
	FillStitch Stitch
}

// ops returns the concrete op list for this round given how many stitches are
// below it (needed to expand a Fill round).
func (r Round) ops(prev int) []Op {
	if r.Fill {
		out := make([]Op, prev)
		op := stitchOp(r.FillStitch)
		for i := range out {
			out[i] = op
		}
		return out
	}
	return r.Ops
}

// height returns the round's representative stitch height (the tallest op).
func (r Round) height() float64 {
	if r.Fill {
		return r.FillStitch.Height()
	}
	h := 0.0
	for _, o := range r.Ops {
		_, _, st := o.spec()
		if sh := st.Height(); sh > h {
			h = sh
		}
	}
	if h == 0 {
		h = 1
	}
	return h
}

// Pattern is a worked-in-the-round crochet pattern: Start stitches in a magic
// ring, then a sequence of rounds.
type Pattern struct {
	Start  int
	Rounds []Round
}

// BuildConfig controls how a Pattern is placed into the world.
type BuildConfig struct {
	Gauge    float64    // world units per stitch
	Center   math3.Vec3 // bottom (magic-ring) centre
	CloseTop bool       // cinch the final round to a top pole (seal for stuffing)
	Yarn     yarn.Config
}

// Build materialises a Pattern into a Fabric. Rounds are stacked along +Y; each
// round's radius follows its stitch count (so increases widen the shape and
// decreases pull it in), and the rise between rounds is set so the slant
// matches the stitch height. Every stitch is linked to the stitch(es) it is
// worked into.
func Build(w *physics.World, p Pattern, cfg BuildConfig) *Fabric {
	f := newFabric(w, cfg.Yarn)
	f.Radial = true
	if cfg.Gauge <= 0 {
		cfg.Gauge = 0.5
	}
	start := p.Start
	if start < 3 {
		start = 3
	}
	width := cfg.Gauge // a single crochet is one gauge unit wide

	center := f.add(cfg.Center, cfg.Yarn.SegmentMass)

	var (
		rings   [][]int
		radii   []float64
		heights []float64
		downs   [][]int // per ring, each stitch's first anchor below
		y       float64
		prevRad float64
	)

	place := func(count int, height float64) []int {
		radius := float64(count) * width / (2 * math.Pi)
		if len(rings) == 0 {
			y = 0
		} else {
			dr := radius - prevRad
			y += math.Sqrt(math.Max(height*height-dr*dr, 0.04*height*height))
		}
		ring := make([]int, count)
		for i := 0; i < count; i++ {
			a := 2 * math.Pi * float64(i) / float64(count)
			pos := cfg.Center.Add(math3.V(radius*math.Cos(a), y, radius*math.Sin(a)))
			ring[i] = f.add(pos, cfg.Yarn.SegmentMass)
		}
		prevRad = radius
		rings = append(rings, ring)
		radii = append(radii, radius)
		heights = append(heights, height)
		return ring
	}

	link := func(a, b int) {
		w.Link(a, b, cfg.Yarn.Stiffness)
		f.Links = append(f.Links, [2]int{a, b})
	}

	// Round 1: the magic ring, every stitch anchored to the centre.
	ring0 := place(start, cfg.Gauge*Single.Height())
	d0 := make([]int, start)
	for i, n := range ring0 {
		link(center, n)
		d0[i] = center
	}
	downs = append(downs, d0)
	f.Strands = append(f.Strands, ringStrand(w, ring0, cfg.Yarn))
	prev := ring0

	// Subsequent rounds.
	for _, rd := range p.Rounds {
		ops := rd.ops(len(prev))

		// Assign anchors to each new stitch by walking the round below.
		var anchors [][]int
		ap := 0
		for _, op := range ops {
			ns, ac, _ := op.spec()
			a := make([]int, 0, ac)
			for k := 0; k < ac && ap < len(prev); k++ {
				a = append(a, prev[ap])
				ap++
			}
			for s := 0; s < ns; s++ {
				anchors = append(anchors, a)
			}
		}

		count := len(anchors)
		if count < 1 {
			break
		}
		ring := place(count, cfg.Gauge*rd.height())
		down := make([]int, count)
		for i := range ring {
			for _, a := range anchors[i] {
				link(a, ring[i])
			}
			if len(anchors[i]) > 0 {
				down[i] = anchors[i][0]
			} else {
				down[i] = -1
			}
		}
		downs = append(downs, down)
		f.Strands = append(f.Strands, ringStrand(w, ring, cfg.Yarn))
		prev = ring
	}

	// Optional top pole so a stuffed shape is sealed.
	if cfg.CloseTop && len(rings) > 0 {
		top := rings[len(rings)-1]
		pole := f.add(cfg.Center.Add(math3.V(0, y+cfg.Gauge*0.5, 0)), cfg.Yarn.SegmentMass)
		for _, n := range top {
			link(pole, n)
		}
	}

	f.buildScriptCells(rings, radii, heights, downs, cfg.Yarn)
	return f
}

// buildScriptCells emits render cells with the true anchor as each stitch's
// downward neighbour.
func (f *Fabric) buildScriptCells(rings [][]int, radii, heights []float64, downs [][]int, cfg yarn.Config) {
	for r := range rings {
		n := len(rings[r])
		if n == 0 {
			continue
		}
		w := 2 * math.Pi * radii[r] / float64(n)
		col := stripeColor(cfg, r)
		for i := 0; i < n; i++ {
			f.Cells = append(f.Cells, StitchCell{
				Node:  rings[r][i],
				Left:  rings[r][(i-1+n)%n],
				Right: rings[r][(i+1)%n],
				Down:  downs[r][i],
				Up:    -1,
				W:     w, H: heights[r], Color: col,
			})
		}
	}
}
