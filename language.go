package main

import "fmt"

const (
	ILLEGAL Token = 0
	EOF           = 1
	WS            = 2

	CONST = 3 // constant literal

	BEGIN = 4 // begin section statement
	EV    = 5 // evaluation section
	EX    = 6 // execution section
	END   = 7 // end section statement

	RDX = 16 // Read X vector and push it on the stack
	RDY = 17 // Read Y vector and push it on the stack
	RDE = 18 // Read total energy and push it on the stack

	PSH = 32 // Push
	POP = 33 // Pop

	CON = 64 // Constant identifier
	REG = 65 // Register identifier

	// comparison
	// x COMP y, where x was pushed before y
	GEQ = 128 // Pushes 1 if x >= y, else 0
	LEQ = 129 // Pushes 1 if x <= y, else 0
	IEQ = 130 // Pushes 1 if x == y, else 0
	GRT = 131 // Pushes 1 if x > y, else 0
	LST = 132 // Pushes 1 if x < y, else 0

	NOT = 256 // Pushes !x
	AND = 257 // Pushes x & y
	IOR = 258 // Pushes x | y
	XOR = 259 // Pushes x ^ y
	ADD = 260 // Pushes x + y
	SUB = 261 // Pushes x - y
	MUL = 262 // Pushes x * y
	DIV = 263 // Pushes x / y, nop if y == 0
	NEG = 264 // Pushes -x

	RID = 512 // Pushes the ID of the first object in current fov
	SCN = 513 // Pop x, y and pushes x, y to first object in current fov
	THR = 514 // Pop and thrust for x units
	TRN = 515 // Pop x, y and turn by the angle given by unit vector with atan(y, x)
	MNE = 516 // Pop and mine with strength x
	REP = 517 // Pop and reproduce using x energy
)

type (
	Token int16

	Instruction interface {
		Int() int16
		Run(*Machine, []int16)
		Parse(*Parser, []int16)
	}
)

func Translate(token Token) Instruction {
	switch token {
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

	case CONST:
		return "CONST"
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
		if program[i] == REG || program[i] == CON {
			constant = true
		}
	}
	return program
}

func runInstruction(m *Machine, code []int16, f func()) {
	if m.pc > len(code)+1 {
		return
	}
	m.pc++
	f()
	if m.pc > len(code)-1 {
		return
	}
	inst := Translate(Token(code[m.pc]))
	inst.Run(m, code)
}

func runInstructionDebug(m *Machine, code []int16, f func()) {
	if m.pc > len(code)+1 {
		return
	}

	fmt.Printf("%v\n", code)
	fmt.Println("pc", m.pc)

	m.pc++
	f()
	fmt.Printf("%v\n\n", m.stack)
	if m.pc > len(code)-1 {
		return
	}
	inst := Translate(Token(code[m.pc]))
	inst.Run(m, code)
}

// instructions
type Illegal int16

func (i Illegal) Int() int16 {
	return int16(i)
}
func (i Illegal) Run(_ *Machine, _ []int16) {}
func (i Illegal) Parse(p *Parser, pr *Program) error {
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
	*program = append(program, e.Int())
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
	*program = append(program, e.Int())
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
	*program = append(program, e.Int())
	tok, lit := p.scanIgnoreWhitespace()
	if tok != CON && tok != REG {
		return fmt.Errorf("unexpected token %s (\"%s\") for PSH", tok, lit)
	}
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

		if len(m.stack) <= 0 {
			return
		}
		switch source.Int() {
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

type Constant int16

func (e Constant) Int() int16 {
	return int16(e)
}
func (e Constant) Run(m *Machine, code []int16) {}

type Register int16

func (e Register) Int() int16 {
	return int16(e)
}
func (e Register) Run(m *Machine, code []int16) {}

type GreaterEqual int16

func (e GreaterEqual) Int() int16 {
	return int16(e)
}
func (e GreaterEqual) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a >= b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
		m.pc += 2
	})
}

type LessEqual int16

func (e LessEqual) Int() int16 {
	return int16(e)
}

func (e LessEqual) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a <= b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
		m.pc += 2
	})
}

type IsEqual int16

func (e IsEqual) Int() int16 {
	return int16(e)
}

func (e IsEqual) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a == b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
		m.pc += 2
	})
}

type GreaterThan int16

func (e GreaterThan) Int() int16 {
	return int16(e)
}

func (e GreaterThan) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a > b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
		m.pc += 2
	})
}

type LessThan int16

func (e LessThan) Int() int16 {
	return int16(e)
}

func (e LessThan) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if a < b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
		m.pc += 2
	})
}

type Not int16

func (e Not) Int() int16 {
	return int16(e)
}

func (e Not) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		if a == 0 {
			m.stack.Push(1)
		} else if a == 1 {
			m.stack.Push(0)
		}
		m.pc++
	})
}

type And int16

func (e And) Int() int16 {
	return int16(e)
}

func (e And) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a & b)
		m.pc += 2
	})
}

type Or int16

func (e Or) Int() int16 {
	return int16(e)
}

func (e Or) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a | b)
		m.pc += 2
	})
}

type Xor int16

func (e Xor) Int() int16 {
	return int16(e)
}

func (e Xor) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a ^ b)
		m.pc += 2
	})
}

type Add int16

func (e Add) Int() int16 {
	return int16(e)
}

func (e Add) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a + b)
		m.pc += 2
	})
}

type Sub int16

func (e Sub) Int() int16 {
	return int16(e)
}

func (e Sub) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a - b)
		m.pc += 2
	})
}

type Mul int16

func (e Mul) Int() int16 {
	return int16(e)
}

func (e Mul) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a * b)
		m.pc += 2
	})
}

type Div int16

func (e Div) Int() int16 {
	return int16(e)
}

func (e Div) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		b, a := m.stack.Pop(), m.stack.Pop()
		if b == 0 {
			return
		}
		m.stack.Push(a / b)
		m.pc += 2
	})
}

type Neg int16

func (n Neg) Int() int16 {
	return int16(n)
}
func (n Neg) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(a * -1)
		m.pc++
	})
}

type RemoteID int16

func (e RemoteID) Int() int16 {
	return int16(e)
}
func (e RemoteID) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(m.state.RemoteID(a))
		m.pc++
	})
}

type Scan int16

func (e Scan) Int() int16 {
	return int16(e)
}

func (e Scan) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(m.state.Scan(a))
		m.pc++
	})
}

type Thrust int16

func (e Thrust) Int() int16 {
	return int16(e)
}

func (e Thrust) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Thrust(a)
		m.pc++
	})
}

type Turn int16

func (e Turn) Int() int16 {
	return int16(e)
}

func (e Turn) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		x, y := m.stack.Pop(), m.stack.Pop()
		m.state.Turn(x, y)
		m.pc++
	})
}

type Mine int16

func (e Mine) Int() int16 {
	return int16(e)
}

func (e Mine) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Mine(a)
		m.pc++
	})
}

type Reproduce int16

func (e Reproduce) Int() int16 {
	return int16(e)
}

func (e Reproduce) Run(m *Machine, code []int16) {
	m.run(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Reproduce(a)
		m.pc++
	})
}
