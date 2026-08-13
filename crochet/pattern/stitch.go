package pattern

// Stitch enumerates the basic crochet stitches. In this engine a stitch is
// primarily a *height*: taller stitches produce taller rows, which is what
// determines the spacing of the physics mesh. The names follow standard
// (US) crochet terminology.
type Stitch int

const (
	Chain      Stitch = iota // ch  — the foundation loop
	Slip                     // sl st — shortest, tight join
	Single                   // sc  — single crochet
	HalfDouble               // hdc — half double crochet
	Double                   // dc  — double crochet
	Treble                   // tr  — treble/triple crochet
)

// Height returns the relative row height of the stitch, in "stitch units".
// A single crochet is the reference at 1.0. These ratios approximate how much
// taller each stitch stands than a single crochet.
func (s Stitch) Height() float64 {
	switch s {
	case Slip:
		return 0.4
	case Single:
		return 1.0
	case HalfDouble:
		return 1.5
	case Double:
		return 2.0
	case Treble:
		return 3.0
	case Chain:
		return 0.6
	default:
		return 1.0
	}
}

// String returns the standard abbreviation.
func (s Stitch) String() string {
	switch s {
	case Chain:
		return "ch"
	case Slip:
		return "sl st"
	case Single:
		return "sc"
	case HalfDouble:
		return "hdc"
	case Double:
		return "dc"
	case Treble:
		return "tr"
	default:
		return "?"
	}
}
