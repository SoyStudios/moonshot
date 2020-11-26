package main

import (
	"fmt"
	"math"
	"strconv"
)

const (
	ILLEGAL Token = 0
	EOF     Token = 1
	WS      Token = 2
	COMMENT Token = 3

	LITERAL Token = 8 // constant literal
	NOP     Token = 9 // no op

	BEGIN Token = 16 // begin section statement
	EV    Token = 17 // evaluation section
	EX    Token = 18 // execution section
	END   Token = 19 // end section statement

	RDX Token = 32 // Read X vector and push it on the stack
	RDY Token = 33 // Read Y vector and push it on the stack
	RDE Token = 34 // Read total energy and push it on the stack

	PSH Token = 64 // Push
	POP Token = 65 // Pop

	CON Token = 128 // Constant identifier
	REG Token = 129 // Register identifier

	// comparison
	// x COMP y, where x was pushed before y
	GEQ Token = 256 // Pushes 1 if x >= y, else 0
	LEQ Token = 257 // Pushes 1 if x <= y, else 0
	IEQ Token = 258 // Pushes 1 if x == y, else 0
	GRT Token = 259 // Pushes 1 if x > y, else 0
	LST Token = 260 // Pushes 1 if x < y, else 0

	NOT Token = 512 // Pushes !x
	AND Token = 513 // Pushes x & y
	IOR Token = 514 // Pushes x | y
	XOR Token = 515 // Pushes x ^ y
	ADD Token = 516 // Pushes x + y
	SUB Token = 517 // Pushes x - y
	MUL Token = 518 // Pushes x * y
	DIV Token = 519 // Pushes x / y, nop if y == 0
	NEG Token = 520 // Pushes -x
	ABS Token = 521 // Pops x and y, and calculates the length of the vector

	RID Token = 1024 // Pushes the ID of the first object in current fov
	SCN Token = 1025 // Pop x, y and pushes x, y to first object in current fov
	THR Token = 1026 // Pop x, y and thrust for the vector
	TRN Token = 1027 // Pop x, y and turn by the angle given by unit vector with atan(y, x)
	MNE Token = 1028 // Pop and mine with strength x
	REP Token = 1029 // Pop and reproduce using x energy
)

type (
	Token int16

	AST []Instruction

	Instruction interface {
		fmt.Stringer
		Run(*Machine, AST)
		Parse(*Parser, *AST) error
	}
)

func Translate(token Token) Instruction {
	switch token {
	case NOP:
		return &Nop{}
	case COMMENT:
		return &Comment{}

	case RDX:
		return &ReadX{}
	case RDY:
		return &ReadY{}
	case RDE:
		return &ReadEnergy{}

	case PSH:
		return &Push{}
	case POP:
		return &Pop{}

	case GEQ:
		return &GreaterEqual{}
	case LEQ:
		return &LessEqual{}
	case IEQ:
		return &IsEqual{}
	case GRT:
		return &GreaterThan{}
	case LST:
		return &LessThan{}

	case NOT:
		return &Not{}
	case AND:
		return &And{}
	case IOR:
		return &Or{}
	case XOR:
		return &Xor{}
	case ADD:
		return &Add{}
	case SUB:
		return &Sub{}
	case MUL:
		return &Mul{}
	case DIV:
		return &Div{}
	case NEG:
		return &Neg{}
	case ABS:
		return &Abs{}

	case RID:
		return &RemoteID{}
	case SCN:
		return &Scan{}
	case THR:
		return &Thrust{}
	case TRN:
		return &Turn{}
	case MNE:
		return &Mine{}
	case REP:
		return &Reproduce{}

	case ILLEGAL:
		fallthrough
	default:
		return Illegal{}
	}
}

func (t Token) String() string {
	switch t {
	case EOF:
		return "EOF"
	case WS:
		return "WS"

	case LITERAL:
		return "LITERAL"
	case NOP:
		return "NOP"
	case COMMENT:
		return "//"

	case BEGIN:
		return "BEGIN"
	case EV:
		return "EV"
	case EX:
		return "EX"
	case END:
		return "END"

	case RDX:
		return "RDX"
	case RDY:
		return "RDY"
	case RDE:
		return "RDE"

	case PSH:
		return "PSH"
	case POP:
		return "POP"

	case CON:
		return "CON"
	case REG:
		return "REG"

	case GEQ:
		return "GEQ"
	case LEQ:
		return "LEQ"
	case IEQ:
		return "IEQ"
	case GRT:
		return "GRT"
	case LST:
		return "LST"

	case NOT:
		return "NOT"
	case AND:
		return "AND"
	case IOR:
		return "IOR"
	case XOR:
		return "XOR"
	case ADD:
		return "ADD"
	case SUB:
		return "SUB"
	case MUL:
		return "MUL"
	case DIV:
		return "DIV"
	case NEG:
		return "NEG"
	case ABS:
		return "ABS"

	case RID:
		return "RID"
	case SCN:
		return "SCN"
	case THR:
		return "THR"
	case TRN:
		return "TRN"
	case MNE:
		return "MNE"
	case REP:
		return "REP"

	case ILLEGAL:
		fallthrough
	default:
		return "ILLEGAL"
	}
}

// instructions

type Illegal struct{}

func (i Illegal) String() string {
	return ILLEGAL.String()
}
func (i Illegal) Run(_ *Machine, _ AST) {}
func (i Illegal) Parse(p *Parser, _ *AST) error {
	p.unscan()
	tok, lit := p.scanIgnoreWhitespace()
	return fmt.Errorf("cannot parse illegal token %s (\"%s\")", tok, lit)
}

type Nop struct{}

func (n Nop) String() string {
	return NOP.String()
}
func (n Nop) Run(m *Machine, code AST) {
	m.run(m, code, func() {})
}
func (n *Nop) Parse(p *Parser, program *AST) error {
	*program = append(*program, n)
	return nil
}

type Comment struct {
	Lit string
}

func (n Comment) String() string {
	return COMMENT.String()
}
func (n Comment) Run(m *Machine, code AST) {
	m.run(m, code, func() {})
}
func (n *Comment) Parse(p *Parser, program *AST) error {
	p.unscan()
	_, n.Lit = p.scanIgnoreWhitespace()
	*program = append(*program, n)
	return nil
}

type ReadX struct{}

func (r ReadX) String() string {
	return RDX.String()
}
func (r ReadX) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		m.stack.Push(m.state.X())
	})
}
func (r *ReadX) Parse(p *Parser, program *AST) error {
	*program = append(*program, r)
	return nil
}

type ReadY struct{}

func (r ReadY) String() string {
	return RDY.String()
}
func (r ReadY) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		m.stack.Push(m.state.Y())
	})
}
func (r *ReadY) Parse(p *Parser, program *AST) error {
	*program = append(*program, r)
	return nil
}

type ReadEnergy struct{}

func (e ReadEnergy) String() string {
	return RDE.String()
}
func (e ReadEnergy) Int() int16 {
	return int16(RDE)
}
func (e ReadEnergy) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		m.stack.Push(m.state.Energy())
	})
}
func (e *ReadEnergy) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Push struct {
	Source Token
	Value  int16
}

func (e Push) String() string {
	return fmt.Sprintf("%s %s %d", PSH, e.Source, e.Value)
}
func (e Push) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		switch e.Source {
		case CON:
			m.stack.Push(e.Value)
		case REG:
			if int(e.Value) > len(m.registers)-1 || e.Value < 0 {
				return
			}
			m.stack.Push(m.registers[e.Value])
		default:
			return
		}
	})
}
func (e *Push) Parse(p *Parser, program *AST) error {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != CON && tok != REG {
		return fmt.Errorf("unexpected token %s (\"%s\") for PSH", tok, lit)
	}
	e.Source = tok
	tok, lit = p.scanIgnoreWhitespace()
	if tok != LITERAL {
		return fmt.Errorf("unexepcted %s (\"%s\") on PSH.", tok, lit)
	}
	v, err := strconv.ParseInt(lit, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid literal: %s. %v", lit, err)
	}
	e.Value = int16(v)
	*program = append(*program, e)
	return nil
}

type Pop struct {
	Index int16
}

func (e Pop) String() string {
	return fmt.Sprintf("%s %s %d", POP, REG, e.Index)
}
func (e Pop) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		m.registers[e.Index] = m.stack.Pop()
	})
}
func (e *Pop) Parse(p *Parser, program *AST) error {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != REG {
		return fmt.Errorf("unexpected token %s (\"%s\") for POP", tok, lit)
	}
	tok, lit = p.scanIgnoreWhitespace()
	if tok != LITERAL {
		return fmt.Errorf("unexpected token %s (\"%s\") for POP. Expect literal.", tok, lit)
	}
	v, err := strconv.ParseInt(lit, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid literal %s: %v", lit, err)
	}
	e.Index = int16(v)
	*program = append(*program, e)

	return nil
}

type GreaterEqual struct{}

func (e GreaterEqual) String() string {
	return GEQ.String()
}
func (e GreaterEqual) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a >= b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}
func (e *GreaterEqual) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type LessEqual struct{}

func (e LessEqual) String() string {
	return LEQ.String()
}
func (e LessEqual) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a <= b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}
func (e *LessEqual) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type IsEqual struct{}

func (e IsEqual) String() string {
	return IEQ.String()
}
func (e IsEqual) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a == b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}
func (e *IsEqual) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type GreaterThan struct{}

func (e GreaterThan) String() string {
	return GRT.String()
}
func (e GreaterThan) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a > b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}
func (e *GreaterThan) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type LessThan struct{}

func (e LessThan) String() string {
	return LST.String()
}
func (e LessThan) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a < b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}
func (e *LessThan) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Not struct{}

func (e Not) String() string {
	return NOT.String()
}
func (e Not) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		if a == 0 {
			m.stack.Push(1)
		} else if a == 1 {
			m.stack.Push(0)
		}
	})
}
func (e *Not) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type And struct{}

func (e And) String() string {
	return AND.String()
}
func (e And) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a & b)
	})
}
func (e *And) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Or struct{}

func (e Or) String() string {
	return IOR.String()
}
func (e Or) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a | b)
	})
}
func (e *Or) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Xor struct{}

func (e Xor) String() string {
	return XOR.String()
}
func (e Xor) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a ^ b)
	})
}
func (e *Xor) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Add struct{}

func (e Add) String() string {
	return ADD.String()
}
func (e Add) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a + b)
	})
}
func (e *Add) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Sub struct{}

func (e Sub) String() string {
	return SUB.String()
}
func (e Sub) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a - b)
	})
}
func (e *Sub) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Mul struct{}

func (e Mul) String() string {
	return MUL.String()
}
func (e Mul) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a * b)
	})
}
func (e *Mul) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Div struct{}

func (e Div) String() string {
	return DIV.String()
}
func (e Div) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if b == 0 {
			return
		}
		m.stack.Push(a / b)
	})
}
func (e *Div) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Neg struct{}

func (n Neg) String() string {
	return NEG.String()
}
func (n Neg) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(a * -1)
	})
}
func (e *Neg) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Abs struct{}

func (n Abs) String() string {
	return ABS.String()
}
func (n Abs) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		x, y := m.stack.Pop(), m.stack.Pop()
		v := int16(math.Round(
			math.Sqrt(
				math.Pow(float64(x), 2) +
					math.Pow(float64(y), 2),
			),
		))
		m.stack.Push(v)
	})
}
func (e *Abs) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type RemoteID struct{}

func (e RemoteID) String() string {
	return RID.String()
}
func (e RemoteID) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(m.state.RemoteID(a))
	})
}
func (e *RemoteID) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Scan struct{}

func (e Scan) String() string {
	return SCN.String()
}
func (e Scan) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		y, x := m.stack.Pop(), m.stack.Pop()
		x, y = m.state.Scan(x, y)
		m.stack.Push(x)
		m.stack.Push(y)
	})
}
func (e *Scan) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Thrust struct{}

func (e Thrust) String() string {
	return THR.String()
}
func (e Thrust) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		y, x := m.stack.Pop(), m.stack.Pop()
		m.state.Thrust(x, y)
	})
}
func (e *Thrust) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Turn struct{}

func (e Turn) String() string {
	return TRN.String()
}
func (e Turn) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Turn(a)
	})
}
func (e *Turn) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Mine struct{}

func (e Mine) String() string {
	return MNE.String()
}
func (e Mine) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Mine(a)
	})
}
func (e *Mine) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}

type Reproduce struct{}

func (e Reproduce) String() string {
	return REP.String()
}
func (e Reproduce) Run(m *Machine, code AST) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Reproduce(a)
	})
}
func (e *Reproduce) Parse(p *Parser, program *AST) error {
	*program = append(*program, e)
	return nil
}
