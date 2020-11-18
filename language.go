package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/pkg/errors"
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
	THR Token = 1026 // Pop and thrust for x units
	TRN Token = 1027 // Pop x, y and turn by the angle given by unit vector with atan(y, x)
	MNE Token = 1028 // Pop and mine with strength x
	REP Token = 1029 // Pop and reproduce using x energy
)

type (
	Token int16

	Instruction interface {
		String(*Machine, []int16) string
		// Int() int16
		Run(*Machine, []int16)
		Parse(*Parser, *[]int16) error
	}
)

func Translate(token Token) Instruction {
	switch token {
	case LITERAL:
		return Literal{}
	case NOP:
		return Nop{}

	case RDX:
		return ReadX{}
	case RDY:
		return ReadY{}
	case RDE:
		return ReadEnergy{}

	case PSH:
		return Push{}
	case POP:
		return Pop{}

	case CON:
		return Constant{}
	case REG:
		return Register{}

	case GEQ:
		return GreaterEqual{}
	case LEQ:
		return LessEqual{}
	case IEQ:
		return IsEqual{}
	case GRT:
		return GreaterThan{}
	case LST:
		return LessThan{}

	case NOT:
		return Not{}
	case AND:
		return And{}
	case IOR:
		return Or{}
	case XOR:
		return Xor{}
	case ADD:
		return Add{}
	case SUB:
		return Sub{}
	case MUL:
		return Mul{}
	case DIV:
		return Div{}
	case NEG:
		return Neg{}
	case ABS:
		return Abs{}

	case RID:
		return RemoteID{}
	case SCN:
		return Scan{}
	case THR:
		return Thrust{}
	case TRN:
		return Turn{}
	case MNE:
		return Mine{}
	case REP:
		return Reproduce{}

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

// TranslateProgram is a utility funtion for converting a
// slice of (readable) tokens into the byte representation
func TranslateProgram(tks []Token) []int16 {
	program := make([]int16, len(tks))
	for i, t := range tks {
		program[i] = int16(t)
	}
	return program
}

// instructions

type Illegal struct{}

func (i Illegal) String(*Machine, []int16) string {
	return ILLEGAL.String()
}
func (i Illegal) Run(_ *Machine, _ []int16) {}
func (i Illegal) Parse(p *Parser, pr *[]int16) error {
	p.unscan()
	tok, lit := p.scanIgnoreWhitespace()
	return fmt.Errorf("cannot parse illegal token %s (\"%s\")", tok, lit)
}

type Literal struct{}

func (e Literal) String(m *Machine, code []int16) string {
	return fmt.Sprintf("literal: %d", code[m.pc])
}
func (e Literal) Run(m *Machine, code []int16) {}
func (e Literal) Parse(p *Parser, program *[]int16) error {
	p.unscan()
	_, lit := p.scanIgnoreWhitespace()
	val, err := strconv.ParseInt(lit, 10, 16)
	if err != nil {
		return errors.Wrap(err, "invalid literal")
	}
	*program = append(*program, int16(val))
	return nil
}

type Nop struct{}

func (n Nop) String(_ *Machine, _ []int16) string {
	return NOP.String()
}
func (n Nop) Run(m *Machine, code []int16) {
	m.run(m, code, func() {})
}
func (n Nop) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, int16(NOP))
	return nil
}

type ReadX struct{}

func (e ReadX) String(*Machine, []int16) string {
	return RDX.String()
}
func (e ReadX) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		m.stack.Push(m.state.X())
	})
}
func (e ReadX) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, int16(RDX))
	return nil
}

type ReadY struct{}

func (e ReadY) String(*Machine, []int16) string {
	return RDY.String()
}
func (e ReadY) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		m.stack.Push(m.state.Y())
	})
}
func (e ReadY) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, int16(RDY))
	return nil
}

type ReadEnergy struct{}

func (e ReadEnergy) String(*Machine, []int16) string {
	return RDE.String()
}
func (e ReadEnergy) Int() int16 {
	return int16(RDE)
}
func (e ReadEnergy) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		m.stack.Push(m.state.Energy())
	})
}
func (e ReadEnergy) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Push struct{}

func (e Push) String(m *Machine, code []int16) string {
	if m.pc+2 > len(code)-1 {
		return "PSH ILLEGAL"
	}
	source := Token(code[m.pc+1])
	sourceValue := code[m.pc+2]
	return fmt.Sprintf("%s %s %d", PSH, source, sourceValue)
}
func (e Push) Int() int16 {
	return int16(PSH)
}
func (e Push) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if m.pc+2 > len(code)-1 {
			return
		}
		source := Token(code[m.pc+1])
		sourceValue := code[m.pc+2]
		m.pc += 2

		switch source {
		case CON:
			m.stack.Push(sourceValue)
		case REG:
			if int(sourceValue) > len(m.registers)-1 || sourceValue < 0 {
				return
			}
			m.stack.Push(m.registers[sourceValue])
		default:
			return
		}
	})
}
func (e Push) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	tok, lit := p.scanIgnoreWhitespace()
	if tok != CON && tok != REG {
		return fmt.Errorf("unexpected token %s (\"%s\") for PSH", tok, lit)
	}
	p.unscan()
	return nil
}

type Pop struct{}

func (e Pop) String(m *Machine, code []int16) string {
	if m.pc+2 > len(code)-1 {
		return "POP ILLEGAL"
	}
	source := Translate(Token(code[m.pc+1]))
	sourceValue := code[m.pc+2]
	return fmt.Sprintf("%s %s %d", POP, source.String(m, code), sourceValue)
}
func (e Pop) Int() int16 {
	return int16(POP)
}
func (e Pop) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if m.pc+2 > len(code)-1 {
			return
		}
		source := Token(code[m.pc+1])
		sourceValue := code[m.pc+2]
		m.pc += 2

		if len(*m.stack) <= 0 {
			return
		}
		switch source {
		case CON:
			return
		case REG:
			if int(sourceValue) > len(m.registers)-1 || sourceValue < 0 {
				return
			}
			m.registers[sourceValue] = m.stack.Pop()
		default:
			return
		}
	})
}
func (e Pop) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	tok, lit := p.scanIgnoreWhitespace()
	if tok != REG {
		return fmt.Errorf("unexpected token %s (\"%s\") for POP", tok, lit)
	}
	p.unscan()
	return nil
}

type Constant struct{}

func (e Constant) String(*Machine, []int16) string {
	return CON.String()
}
func (e Constant) Int() int16 {
	return int16(CON)
}
func (e Constant) Run(m *Machine, code []int16) {}
func (e Constant) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	tok, lit := p.scanIgnoreWhitespace()
	if tok != LITERAL {
		return fmt.Errorf("unexpected token %s (\"%s\") expecting LITERAL", tok, lit)
	}
	p.unscan()
	return nil
}

type Register struct{}

func (e Register) String(*Machine, []int16) string {
	return REG.String()
}
func (e Register) Int() int16 {
	return int16(REG)
}
func (e Register) Run(m *Machine, code []int16) {}
func (e Register) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	tok, lit := p.scanIgnoreWhitespace()
	if tok != LITERAL {
		return fmt.Errorf("unexpected token %s (\"%s\") expecting CONST", tok, lit)
	}
	p.unscan()
	return nil
}

type GreaterEqual struct{}

func (e GreaterEqual) String(*Machine, []int16) string {
	return GEQ.String()
}
func (e GreaterEqual) Int() int16 {
	return int16(GEQ)
}
func (e GreaterEqual) Run(m *Machine, code []int16) {
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
func (e GreaterEqual) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type LessEqual struct{}

func (e LessEqual) String(*Machine, []int16) string {
	return LEQ.String()
}
func (e LessEqual) Int() int16 {
	return int16(LEQ)
}
func (e LessEqual) Run(m *Machine, code []int16) {
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
func (e LessEqual) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type IsEqual struct{}

func (e IsEqual) String(*Machine, []int16) string {
	return IEQ.String()
}
func (e IsEqual) Int() int16 {
	return int16(IEQ)
}
func (e IsEqual) Run(m *Machine, code []int16) {
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
func (e IsEqual) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type GreaterThan struct{}

func (e GreaterThan) String(*Machine, []int16) string {
	return GRT.String()
}
func (e GreaterThan) Int() int16 {
	return int16(GRT)
}
func (e GreaterThan) Run(m *Machine, code []int16) {
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
func (e GreaterThan) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type LessThan struct{}

func (e LessThan) String(*Machine, []int16) string {
	return LST.String()
}
func (e LessThan) Int() int16 {
	return int16(LST)
}
func (e LessThan) Run(m *Machine, code []int16) {
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
func (e LessThan) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Not struct{}

func (e Not) String(*Machine, []int16) string {
	return NOT.String()
}
func (e Not) Int() int16 {
	return int16(NOT)
}
func (e Not) Run(m *Machine, code []int16) {
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
func (e Not) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type And struct{}

func (e And) String(*Machine, []int16) string {
	return AND.String()
}
func (e And) Int() int16 {
	return int16(AND)
}
func (e And) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a & b)
	})
}
func (e And) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Or struct{}

func (e Or) String(*Machine, []int16) string {
	return IOR.String()
}
func (e Or) Int() int16 {
	return int16(IOR)
}
func (e Or) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a | b)
	})
}
func (e Or) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Xor struct{}

func (e Xor) String(*Machine, []int16) string {
	return XOR.String()
}
func (e Xor) Int() int16 {
	return int16(XOR)
}
func (e Xor) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a ^ b)
	})
}
func (e Xor) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Add struct{}

func (e Add) String(*Machine, []int16) string {
	return ADD.String()
}
func (e Add) Int() int16 {
	return int16(ADD)
}
func (e Add) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a + b)
	})
}
func (e Add) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Sub struct{}

func (e Sub) String(*Machine, []int16) string {
	return SUB.String()
}
func (e Sub) Int() int16 {
	return int16(SUB)
}
func (e Sub) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a - b)
	})
}
func (e Sub) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Mul struct{}

func (e Mul) String(*Machine, []int16) string {
	return MUL.String()
}
func (e Mul) Int() int16 {
	return int16(MUL)
}
func (e Mul) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a * b)
	})
}
func (e Mul) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Div struct{}

func (e Div) String(*Machine, []int16) string {
	return DIV.String()
}
func (e Div) Int() int16 {
	return int16(DIV)
}
func (e Div) Run(m *Machine, code []int16) {
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
func (e Div) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Neg struct{}

func (n Neg) String(*Machine, []int16) string {
	return NEG.String()
}
func (n Neg) Int() int16 {
	return int16(NEG)
}
func (n Neg) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(a * -1)
	})
}
func (e Neg) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Abs struct{}

func (n Abs) String(*Machine, []int16) string {
	return ABS.String()
}
func (n Abs) Int() int16 {
	return int16(ABS)
}
func (n Abs) Run(m *Machine, code []int16) {
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
func (e Abs) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type RemoteID struct{}

func (e RemoteID) String(*Machine, []int16) string {
	return RID.String()
}
func (e RemoteID) Int() int16 {
	return int16(RID)
}
func (e RemoteID) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(m.state.RemoteID(a))
	})
}
func (e RemoteID) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Scan struct{}

func (e Scan) String(*Machine, []int16) string {
	return SCN.String()
}
func (e Scan) Int() int16 {
	return int16(SCN)
}
func (e Scan) Run(m *Machine, code []int16) {
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
func (e Scan) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Thrust struct{}

func (e Thrust) String(*Machine, []int16) string {
	return THR.String()
}
func (e Thrust) Int() int16 {
	return int16(THR)
}
func (e Thrust) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Thrust(a)
	})
}
func (e Thrust) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Turn struct{}

func (e Turn) String(*Machine, []int16) string {
	return TRN.String()
}
func (e Turn) Int() int16 {
	return int16(TRN)
}
func (e Turn) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		y, x := m.stack.Pop(), m.stack.Pop()
		m.state.Turn(x, y)
	})
}
func (e Turn) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Mine struct{}

func (e Mine) String(*Machine, []int16) string {
	return MNE.String()
}
func (e Mine) Int() int16 {
	return int16(MNE)
}
func (e Mine) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Mine(a)
	})
}
func (e Mine) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Reproduce struct{}

func (e Reproduce) String(*Machine, []int16) string {
	return REP.String()
}
func (e Reproduce) Int() int16 {
	return int16(REP)
}
func (e Reproduce) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Reproduce(a)
	})
}
func (e Reproduce) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}
