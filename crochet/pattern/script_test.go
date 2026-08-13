package pattern

import (
	"testing"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

func TestParseStartCount(t *testing.T) {
	p, err := Parse("R1: 6 sc in magic ring")
	if err != nil {
		t.Fatal(err)
	}
	if p.Start != 6 {
		t.Fatalf("Start = %d, want 6", p.Start)
	}
	if len(p.Rounds) != 0 {
		t.Fatalf("expected no further rounds, got %d", len(p.Rounds))
	}
}

func TestParseIncAndRepeat(t *testing.T) {
	p := MustParse(`
		R1: 6 sc in magic ring
		R2: inc x6
		R3: (sc, inc) x6
		R4: (2 sc, inc) x6
	`)
	if p.Start != 6 {
		t.Fatalf("Start = %d", p.Start)
	}
	if len(p.Rounds) != 3 {
		t.Fatalf("rounds = %d, want 3", len(p.Rounds))
	}
	// The resulting stitch counts should be 12, 18, 24.
	prev := p.Start
	want := []int{12, 18, 24}
	for i, rd := range p.Rounds {
		got := countNewStitches(rd.ops(prev))
		if got != want[i] {
			t.Fatalf("round %d count = %d, want %d", i+2, got, want[i])
		}
		prev = got
	}
}

func TestParseFillAndRange(t *testing.T) {
	p := MustParse(`
		R1: 6 sc in magic ring
		R2: inc x6
		R3-5: sc
	`)
	if len(p.Rounds) != 4 { // one inc round + three fill rounds
		t.Fatalf("rounds = %d, want 4", len(p.Rounds))
	}
	for i := 1; i < 4; i++ {
		if !p.Rounds[i].Fill {
			t.Fatalf("round index %d should be a fill round", i)
		}
	}
	// A fill round keeps the count the same (12 sc into 12).
	if got := countNewStitches(p.Rounds[1].ops(12)); got != 12 {
		t.Fatalf("fill round count = %d, want 12", got)
	}
}

func TestParseDecreaseForms(t *testing.T) {
	p := MustParse(`
		6 sc in magic ring
		(sc, dec) x2
		sc2tog x3
	`)
	// (sc, dec) x2 over 6 below → 2*(1+1)=4 stitches, consuming 2*(1+2)=6.
	if got := countNewStitches(p.Rounds[0].ops(6)); got != 4 {
		t.Fatalf("(sc,dec)x2 count = %d, want 4", got)
	}
	// sc2tog x3 → 3 stitches (each a decrease).
	for _, o := range p.Rounds[1].ops(6) {
		if o != OpDec {
			t.Fatalf("sc2tog should parse as OpDec, got %v", o)
		}
	}
}

func TestBuildStitchCounts(t *testing.T) {
	p := MustParse(`
		R1: 6 sc in magic ring
		R2: inc x6
		R3: (sc, inc) x6
	`)
	w := physics.NewWorld()
	f := Build(w, p, BuildConfig{Gauge: 0.5, Center: math3.V(0, 5, 0), Yarn: yarn.DefaultConfig()})

	// 6 + 12 + 18 stitches + 1 centre node = 37 particles/cells+centre.
	if got := len(f.Cells); got != 6+12+18 {
		t.Fatalf("cells = %d, want 36", got)
	}
	if got := len(f.Nodes); got != 6+12+18+1 {
		t.Fatalf("nodes = %d, want 37 (incl centre)", got)
	}
	// Every non-magic-ring cell must anchor to a real stitch below.
	below := 0
	for _, c := range f.Cells {
		if c.Down >= 0 {
			below++
		}
	}
	if below != len(f.Cells) {
		t.Fatalf("some cells have no anchor below: %d/%d", below, len(f.Cells))
	}
}

func TestBuiltAmigurumiStaysFinite(t *testing.T) {
	p := MustParse(`
		R1: 6 sc in magic ring
		R2: inc x6
		R3: (sc, inc) x6
		R4: (2 sc, inc) x6
		R5-6: sc
		R7: (2 sc, dec) x6
		R8: (sc, dec) x6
		R9: dec x6
	`)
	w := physics.NewWorld()
	w.Gravity = math3.Zero
	w.SelfCollision = true
	w.CollisionRadius = 0.14
	f := Build(w, p, BuildConfig{Gauge: 0.45, Center: math3.V(0, 5, 0), CloseTop: true, Yarn: yarn.DefaultConfig()})
	f.Stuff(0.004)

	for n := 0; n < 300; n++ {
		w.Step(1.0 / 120)
	}
	for i, pt := range w.Particles {
		if math3IsBad(pt.Pos) {
			t.Fatalf("particle %d non-finite: %+v", i, pt.Pos)
		}
	}
}

func math3IsBad(v math3.Vec3) bool {
	return v != v || // NaN
		v.X > 1e9 || v.X < -1e9 || v.Y > 1e9 || v.Y < -1e9 || v.Z > 1e9 || v.Z < -1e9
}
