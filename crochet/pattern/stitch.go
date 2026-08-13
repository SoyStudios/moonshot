package pattern

// Stitch enumerates the basic crochet stitches. Each stitch carries a physical
// "gauge": how wide it sits and how tall it stands, in stitch units where a
// single crochet is 1×1. Taller stitches (dc, tr) make taller rows; that gauge
// drives the spacing of the physics mesh the builders create. Names follow
// standard US crochet terminology.
type Stitch int

const (
	Chain      Stitch = iota // ch  — the foundation loop
	Slip                     // sl st — shortest, tight join
	Single                   // sc  — single crochet
	HalfDouble               // hdc — half double crochet
	Double                   // dc  — double crochet
	Treble                   // tr  — treble/triple crochet
)

// Def describes the gauge and appearance of a stitch.
type Def struct {
	Abbr   string
	Width  float64 // horizontal footprint in stitch units
	Height float64 // row height in stitch units
}

// Def returns the stitch's gauge definition.
func (s Stitch) Def() Def {
	switch s {
	case Chain:
		return Def{"ch", 0.9, 0.6}
	case Slip:
		return Def{"sl st", 1.0, 0.4}
	case Single:
		return Def{"sc", 1.0, 1.0}
	case HalfDouble:
		return Def{"hdc", 1.05, 1.5}
	case Double:
		return Def{"dc", 1.1, 2.0}
	case Treble:
		return Def{"tr", 1.15, 3.0}
	default:
		return Def{"sc", 1.0, 1.0}
	}
}

// Height returns the relative row height of the stitch (single crochet = 1.0).
func (s Stitch) Height() float64 { return s.Def().Height }

// Width returns the relative horizontal footprint of the stitch.
func (s Stitch) Width() float64 { return s.Def().Width }

// String returns the standard abbreviation.
func (s Stitch) String() string { return s.Def().Abbr }
