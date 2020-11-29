package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

var eof = rune(0)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || ch == '-'
}

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	// nolint: errcheck
	s.r.UnreadRune()
}

func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) || isDigit(ch) {
		s.unread()
		return s.scanIdent()
	}

	switch ch {
	case eof:
		return EOF, ""
	case '/':
		return s.scanComment()
	default:
		return ILLEGAL, string(ch)
	}
}

func (s *Scanner) scanComment() (tok Token, lit string) {
	s.unread()
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if ch == '\n' {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	s.scanWhitespace()

	return COMMENT, buf.String()
}

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

func (s *Scanner) scanIdent() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	switch strings.ToUpper(buf.String()) {
	case "BEGIN":
		return BEGIN, buf.String()
	case "EV":
		return EV, buf.String()
	case "EX":
		return EX, buf.String()
	case "END":
		return END, buf.String()

	case "NOP":
		return NOP, buf.String()

	case "RDX":
		return RDX, buf.String()
	case "RDY":
		return RDY, buf.String()
	case "RDE":
		return RDE, buf.String()

	case "PSH":
		return PSH, buf.String()
	case "POP":
		return POP, buf.String()

	case "CON":
		return CON, buf.String()
	case "REG":
		return REG, buf.String()
	case "RMT":
		return RMT, buf.String()

	case "GEQ":
		return GEQ, buf.String()
	case "LEQ":
		return LEQ, buf.String()
	case "IEQ":
		return IEQ, buf.String()
	case "GRT":
		return GRT, buf.String()
	case "LST":
		return LST, buf.String()

	case "NOT":
		return NOT, buf.String()
	case "AND":
		return AND, buf.String()
	case "IOR":
		return IOR, buf.String()
	case "XOR":
		return XOR, buf.String()
	case "ADD":
		return ADD, buf.String()
	case "SUB":
		return SUB, buf.String()
	case "MUL":
		return MUL, buf.String()
	case "DIV":
		return DIV, buf.String()
	case "NEG":
		return NEG, buf.String()
	case "ABS":
		return ABS, buf.String()

	case "RID":
		return RID, buf.String()
	case "SCN":
		return SCN, buf.String()
	case "THR":
		return THR, buf.String()
	case "TRN":
		return TRN, buf.String()
	case "MNE":
		return MNE, buf.String()
	case "REP":
		return REP, buf.String()
	case "IMP":
		return IMP, buf.String()

	default:
		return LITERAL, buf.String()
	}
}

type Parser struct {
	s   *Scanner
	buf struct {
		tok Token
		lit string
		n   int
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) scan() (tok Token, lit string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	tok, lit = p.s.Scan()
	p.buf.tok, p.buf.lit = tok, lit
	return
}
func (p *Parser) unscan() { p.buf.n = 1 }
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) parseSection(section *AST, require Token) error {
	for {
		tok, _ := p.scanIgnoreWhitespace()
		if tok == EOF {
			return fmt.Errorf("unexpected %s in evaluation section", EOF)
		}
		if tok == END {
			break
		}
		inst := Translate(tok)
		err := inst.Parse(p, section)
		if err != nil {
			return err
		}
	}
	// section can start with comments and must start with a BEGIN
	for _, inst := range *section {
		_, ok := inst.(*Comment)
		if ok {
			continue
		}
		begin, ok := inst.(*Begin)
		if !ok {
			return fmt.Errorf("unexpected instruction %s. Expect %s", inst, BEGIN)
		}
		if begin.Section != require {
			return fmt.Errorf("unexpected section %s. Expect %s.", begin.Section, require)
		}
		break
	}
	return nil
}

func (p *Parser) Parse() ([]*Gene, error) {
	pr := make([]*Gene, 0)
	for {
		if tok, _ := p.scanIgnoreWhitespace(); tok == EOF {
			break
		}
		p.unscan()

		g := NewGene()

		err := p.parseSection(&g.Evaluate, EV)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing evaluation section")
		}
		err = p.parseSection(&g.Execute, EX)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing execution section")
		}
		pr = append(pr, g)
	}
	return pr, nil
}
