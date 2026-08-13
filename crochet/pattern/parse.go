package pattern

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Parse reads a worked-in-the-round crochet pattern in standard written
// notation and returns a Pattern. It understands the common forms:
//
//	R1: 6 sc in magic ring
//	R2: inc x6
//	R3: (sc, inc) x6
//	R4: (2 sc, inc) x6
//	R5-7: sc                // "sc in each st around"
//	R8: (2 sc, dec) x6
//	R9: sc2tog x6
//
// Round labels (R1:, Rnd 1:, Round 1:, 1)) are optional; rounds may also be
// given one per line with no label. A trailing stitch-count like "(18)" and
// filler such as "in each st" / "around" are ignored. Repeats use x, *, or ×.
// The first round is taken as the magic-ring start (its stitch count).
func Parse(text string) (Pattern, error) {
	var p Pattern
	first := true

	for _, raw := range strings.Split(text, "\n") {
		line := stripComment(raw)
		lo, hi, body := splitLabel(line)
		body = cleanup(body)
		if body == "" {
			continue
		}

		if first {
			ops, err := parseBody(body)
			if err != nil {
				return p, fmt.Errorf("round 1 (magic ring): %w", err)
			}
			p.Start = countNewStitches(ops)
			first = false
			continue
		}

		var round Round
		if st, ok := loneStitch(body); ok {
			round = Round{Fill: true, FillStitch: st}
		} else {
			ops, err := parseBody(body)
			if err != nil {
				return p, fmt.Errorf("round %q: %w", body, err)
			}
			round = Round{Ops: ops}
		}

		reps := 1
		if hi >= lo && lo > 0 {
			reps = hi - lo + 1
		}
		for i := 0; i < reps; i++ {
			p.Rounds = append(p.Rounds, round)
		}
	}

	if first {
		return p, fmt.Errorf("empty pattern")
	}
	return p, nil
}

// MustParse is Parse but panics on error — handy for literal patterns in code.
func MustParse(text string) Pattern {
	p, err := Parse(text)
	if err != nil {
		panic(err)
	}
	return p
}

var (
	labelRe = regexp.MustCompile(`(?i)^\s*(?:r(?:ou)?nd?|r)\s*(\d+)\s*(?:[-–]\s*(\d+))?\s*[:.)\-]\s*(.*)$`)
	annRe   = regexp.MustCompile(`\([\s\d]+\)`)             // "(18)" stitch-count notes
	multRe  = regexp.MustCompile(`(?i)[x*×]\s*(\d+)\s*$`)   // "x6", "*6", "×6"
	leadRe  = regexp.MustCompile(`^\s*(\d+)\s+(.*)$`)       // "2 sc"
	fillRe  = regexp.MustCompile(`\b(in each stitch|in each st|in each|around|in the round)\b`)
	ringRe  = regexp.MustCompile(`\b(in )?(a )?(magic ring|magic circle|mr|ring)\b`)
)

func stripComment(s string) string {
	if i := strings.IndexByte(s, '#'); i >= 0 {
		s = s[:i]
	}
	if i := strings.Index(s, "//"); i >= 0 {
		s = s[:i]
	}
	return s
}

// splitLabel pulls an optional round label / range off the front of a line.
func splitLabel(line string) (lo, hi int, body string) {
	if m := labelRe.FindStringSubmatch(line); m != nil {
		lo, _ = strconv.Atoi(m[1])
		hi = lo
		if m[2] != "" {
			hi, _ = strconv.Atoi(m[2])
		}
		return lo, hi, m[3]
	}
	return 0, 0, line
}

func cleanup(s string) string {
	s = strings.ToLower(s)
	s = annRe.ReplaceAllString(s, " ")
	s = ringRe.ReplaceAllString(s, " ")
	s = fillRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// loneStitch reports whether the whole body is a single bare stitch word,
// meaning "one of these in every stitch" (a fill round).
func loneStitch(body string) (Stitch, bool) {
	switch strings.TrimSpace(body) {
	case "sc", "single", "single crochet":
		return Single, true
	case "hdc":
		return HalfDouble, true
	case "dc", "double":
		return Double, true
	case "tr", "treble", "triple":
		return Treble, true
	case "sl st", "slst", "slip", "sl":
		return Slip, true
	}
	return 0, false
}

func parseBody(s string) ([]Op, error) {
	var ops []Op
	for _, seg := range splitTopLevel(s) {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		segOps, err := parseSegment(seg)
		if err != nil {
			return nil, err
		}
		ops = append(ops, segOps...)
	}
	return ops, nil
}

// splitTopLevel splits on commas that are not inside parentheses.
func splitTopLevel(s string) []string {
	var out []string
	depth, start := 0, 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	return append(out, s[start:])
}

func parseSegment(seg string) ([]Op, error) {
	// A parenthesised group, optionally repeated: "(sc, inc) x6".
	if strings.HasPrefix(seg, "(") {
		close := matchParen(seg)
		if close < 0 {
			return nil, fmt.Errorf("unbalanced parentheses in %q", seg)
		}
		inner := seg[1:close]
		mult := extractMult(seg[close+1:])
		group, err := parseBody(inner)
		if err != nil {
			return nil, err
		}
		return repeat(group, mult), nil
	}

	// A plain term: "[N] <stitch> [xM]".
	mult := extractMult(seg)
	seg = multRe.ReplaceAllString(seg, "")
	count := 1
	if m := leadRe.FindStringSubmatch(seg); m != nil {
		count, _ = strconv.Atoi(m[1])
		seg = m[2]
	}
	op, ok := stitchWord(strings.TrimSpace(seg))
	if !ok {
		return nil, fmt.Errorf("unknown stitch %q", strings.TrimSpace(seg))
	}
	return repeat([]Op{op}, count*mult), nil
}

func stitchWord(s string) (Op, bool) {
	switch strings.TrimSpace(s) {
	case "sc", "single", "single crochet":
		return OpSc, true
	case "inc", "increase", "incr":
		return OpInc, true
	case "dec", "decrease", "sc2tog", "dec2tog", "invdec", "invisible decrease":
		return OpDec, true
	case "hdc":
		return OpHdc, true
	case "dc", "double", "double crochet":
		return OpDc, true
	case "tr", "treble", "triple", "trc":
		return OpTr, true
	case "sl st", "slst", "slip", "sl", "slip stitch":
		return OpSlSt, true
	case "ch", "chain": // foundation chain — approximated as sc for counting
		return OpSc, true
	}
	return 0, false
}

// matchParen returns the index of the ')' matching the '(' at position 0.
func matchParen(s string) int {
	depth := 0
	for i, r := range s {
		if r == '(' {
			depth++
		} else if r == ')' {
			if depth--; depth == 0 {
				return i
			}
		}
	}
	return -1
}

func extractMult(s string) int {
	if m := multRe.FindStringSubmatch(s); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

func repeat(ops []Op, n int) []Op {
	if n <= 1 {
		return ops
	}
	out := make([]Op, 0, len(ops)*n)
	for i := 0; i < n; i++ {
		out = append(out, ops...)
	}
	return out
}

func countNewStitches(ops []Op) int {
	total := 0
	for _, o := range ops {
		ns, _, _ := o.spec()
		total += ns
	}
	return total
}
