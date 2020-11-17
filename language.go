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
		Int() int16
		Run(*Machine, []int16)
		Parse(*Parser, *[]int16) error
	}
)

func Translate(token Token) Instruction {
	switch token {
	case LITERAL:
		return Literal(LITERAL)

	case RDX:
		return ReadX(RDX)
	case RDY:
		return ReadY(RDY)
	case RDE:
		return ReadEnergy(RDE)

	case PSH:
		return Push(PSH)
	case POP:
		return Pop(POP)

	case CON:
		return Constant(CON)
	case REG:
		return Register(REG)

	case GEQ:
		return GreaterEqual(GEQ)
	case LEQ:
		return LessEqual(LEQ)
	case IEQ:
		return IsEqual(IEQ)
	case GRT:
		return GreaterThan(GRT)
	case LST:
		return LessThan(LST)

	case NOT:
		return Not(NOT)
	case AND:
		return And(AND)
	case IOR:
		return Or(IOR)
	case XOR:
		return Xor(XOR)
	case ADD:
		return Add(ADD)
	case SUB:
		return Sub(SUB)
	case MUL:
		return Mul(MUL)
	case DIV:
		return Div(DIV)
	case NEG:
		return Neg(NEG)
	case ABS:
		return Abs(ABS)

	case RID:
		return RemoteID(RID)
	case SCN:
		return Scan(SCN)
	case THR:
		return Thrust(THR)
	case TRN:
		return Turn(TRN)
	case MNE:
		return Mine(MNE)
	case REP:
		return Reproduce(REP)

	case ILLEGAL:
		fallthrough
	default:
		return Illegal(ILLEGAL)
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

func TranslateProgram(tks []Token) []int16 {
	program := make([]int16, len(tks))
	var constant bool
	for i, t := range tks {
		if constant {
			constant = false
			program[i] = int16(t)
			continue
		}
		program[i] = Translate(t).Int()
		if program[i] == int16(REG) || program[i] == int16(CON) {
			constant = true
		}
	}
	return program
}

// instructions
type Illegal int16

func (i Illegal) Int() int16 {
	return int16(i)
}
func (i Illegal) Run(_ *Machine, _ []int16) {}
func (i Illegal) Parse(p *Parser, pr *[]int16) error {
	p.unscan()
	tok, lit := p.scanIgnoreWhitespace()
	return fmt.Errorf("cannot parse illegal token %s (\"%s\")", tok, lit)
}

type ReadX int16

func (e ReadX) Int() int16 {
	return int16(e)
}
func (e ReadX) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		m.stack.Push(m.state.X())
	})
}
func (e ReadX) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type ReadY int16

func (e ReadY) Int() int16 {
	return int16(e)
}
func (e ReadY) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		m.stack.Push(m.state.Y())
	})
}
func (e ReadY) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type ReadEnergy int16

func (e ReadEnergy) Int() int16 {
	return int16(e)
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

type Push int16

func (e Push) Int() int16 {
	return int16(e)
}
func (e Push) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if m.pc+2 > len(code)-1 {
			return
		}
		source := Translate(Token(code[m.pc]))
		sourceValue := code[m.pc+1]
		m.pc += 2

		switch source.Int() {
		case int16(CON):
			m.stack.Push(sourceValue)
		case int16(REG):
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

type Pop int16

func (e Pop) Int() int16 {
	return int16(e)
}
func (e Pop) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if m.pc+2 > len(code)-1 {
			return
		}
		source := Translate(Token(code[m.pc]))
		sourceValue := code[m.pc+1]
		m.pc += 2

		if len(*m.stack) <= 0 {
			return
		}
		switch source.Int() {
		case int16(CON):
			return
		case int16(REG):
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

type Literal int16

func (e Literal) Int() int16 {
	return int16(e)
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

type Constant int16

func (e Constant) Int() int16 {
	return int16(e)
}
func (e Constant) Run(m *Machine, code []int16) {}
func (e Constant) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	tok, lit := p.scanIgnoreWhitespace()
	if tok != LITERAL {
		return fmt.Errorf("unexpected token %s (\"%s\") expecting CONST", tok, lit)
	}
	p.unscan()
	return nil
}

type Register int16

func (e Register) Int() int16 {
	return int16(e)
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

type GreaterEqual int16

func (e GreaterEqual) Int() int16 {
	return int16(e)
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

type LessEqual int16

func (e LessEqual) Int() int16 {
	return int16(e)
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

type IsEqual int16

func (e IsEqual) Int() int16 {
	return int16(e)
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

type GreaterThan int16

func (e GreaterThan) Int() int16 {
	return int16(e)
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

type LessThan int16

func (e LessThan) Int() int16 {
	return int16(e)
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

type Not int16

func (e Not) Int() int16 {
	return int16(e)
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

type And int16

func (e And) Int() int16 {
	return int16(e)
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

type Or int16

func (e Or) Int() int16 {
	return int16(e)
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

type Xor int16

func (e Xor) Int() int16 {
	return int16(e)
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

type Add int16

func (e Add) Int() int16 {
	return int16(e)
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

type Sub int16

func (e Sub) Int() int16 {
	return int16(e)
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

type Mul int16

func (e Mul) Int() int16 {
	return int16(e)
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

type Div int16

func (e Div) Int() int16 {
	return int16(e)
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

type Neg int16

func (n Neg) Int() int16 {
	return int16(n)
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

type Abs int16

func (n Abs) Int() int16 {
	return int16(n)
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

type RemoteID int16

func (e RemoteID) Int() int16 {
	return int16(e)
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

type Scan int16

func (e Scan) Int() int16 {
	return int16(e)
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

type Thrust int16

func (e Thrust) Int() int16 {
	return int16(e)
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

type Turn int16

func (e Turn) Int() int16 {
	return int16(e)
}
func (e Turn) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(*m.stack) <= 0 {
			return
		}
		x, y := m.stack.Pop(), m.stack.Pop()
		m.state.Turn(x, y)
	})
}
func (e Turn) Parse(p *Parser, program *[]int16) error {
	*program = append(*program, e.Int())
	return nil
}

type Mine int16

func (e Mine) Int() int16 {
	return int16(e)
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

type Reproduce int16

func (e Reproduce) Int() int16 {
	return int16(e)
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
