package main

const (
	ILLEGAL Token = 0
	EOF           = 1
	WS            = 2

	CONST = 3

	RDX = 7  // Read X vector and push it on the stack
	RDY = 8  // Read Y vector and push it on the stack
	RDA = 9  // Read angle and push it on the stack
	RDE = 10 // Read total energy and push it on the stack

	PSH = 11 // Push
	POP = 12 // Pop

	CON = 13 // Constant identifier
	REG = 14 // Register identifier

	// comparison
	// x COMP y, where x was pushed before y
	GEQ = 17 // Pushes 1 if x >= y, else 0
	LEQ = 18 // Pushes 1 if x <= y, else 0
	IEQ = 19 // Pushes 1 if x == y, else 0
	GRT = 20
	LST = 21

	NOT = 22
	AND = 23
	OR  = 24
	XOR = 25
	ADD = 26
	SUB = 27
	MUL = 28
	DIV = 29

	RID
	SCN
	THR = 20 // Pop and thrust for n units
	TRN = 21 // Turn by n degrees
	MIN = 22
	REP = 23
)

type (
	Token int16

	Instruction interface {
		Int() int16
		Run(*Machine, []int16)
	}
)

func Translate(token Token) Instruction {
	switch token {
	case ILLEGAL:
		fallthrough
	case RDX:
		return ReadX(RDX)
	case RDY:
		return ReadY(RDY)
	default:
		return Illegal(ILLEGAL)
	}
}

func TranslateProgram(tks []Token) []int16 {
	program := make([]int16, len(tks))
	for i, t := range tks {
		program[i] = Translate(t).Int()
	}
	return program
}

func runInstruction(m *Machine, code []int16, f func()) {
	if m.pc > len(code)+1 {
		return
	}
	m.pc++
	f()
	inst := Translate(Token(code[m.pc]))
	inst.Run(m, code)
}

// instructions
type Illegal int16

func (i Illegal) Int() int16 {
	return int16(i)
}

func (i Illegal) Run(_ *Machine, _ []int16) {}

type ReadX int16

func (e ReadX) Int() int16 {
	return int16(e)
}

func (e ReadX) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		m.stack.Push(m.state.X())
	})
}

type ReadY int16

func (e ReadY) Int() int16 {
	return int16(e)
}

func (e ReadY) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		m.stack.Push(m.state.Y())
	})
}

type Push int16

func (e Push) Int() int16 {
	return int16(e)
}

func (e Push) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if m.pc+2 > len(code)+1 {
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

type Pop int16

func (e Pop) Int() int16 {
	return int16(e)
}

func (e Pop) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if m.pc+2 > len(code)+1 {
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

type GreaterEqual int16

func (e GreaterEqual) Int() int16 {
	return int16(e)
}

func (e GreaterEqual) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		if a >= b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}

type LessEqual int16

func (e LessEqual) Int() int16 {
	return int16(e)
}

func (e LessEqual) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		if a <= b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}

type IsEqual int16

func (e IsEqual) Int() int16 {
	return int16(e)
}

func (e IsEqual) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		if a == b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}

type GreaterThan int16

func (e GreaterThan) Int() int16 {
	return int16(e)
}

func (e GreaterThan) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		if a > b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}

type LessThan int16

func (e LessThan) Int() int16 {
	return int16(e)
}

func (e LessThan) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		if a < b {
			m.stack.Push(1)
		} else {
			m.stack.Push(0)
		}
	})
}

type Not int16

func (e Not) Int() int16 {
	return int16(e)
}

func (e Not) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 0 {
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

type And int16

func (e And) Int() int16 {
	return int16(e)
}

func (e And) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a & b)
	})
}

type Or int16

func (e Or) Int() int16 {
	return int16(e)
}

func (e Or) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a | b)
	})
}

type Xor int16

func (e Xor) Int() int16 {
	return int16(e)
}

func (e Xor) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a ^ b)
	})
}

type Add int16

func (e Add) Int() int16 {
	return int16(e)
}

func (e Add) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a + b)
	})
}

type Sub int16

func (e Sub) Int() int16 {
	return int16(e)
}

func (e Sub) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a - b)
	})
}

type Mul int16

func (e Mul) Int() int16 {
	return int16(e)
}

func (e Mul) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a * b)
	})
}

type Div int16

func (e Div) Int() int16 {
	return int16(e)
}

func (e Div) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 1 {
			return
		}
		a, b := m.stack.Pop(), m.stack.Pop()
		m.stack.Push(a / b)
	})
}

type RemoteID int16

func (e RemoteID) Int() int16 {
	return int16(e)
}

func (e RemoteID) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(m.state.RemoteID(a))
	})
}

type Scan int16

func (e Scan) Int() int16 {
	return int16(e)
}

func (e Scan) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.stack.Push(m.state.Scan(a))
	})
}

type Thrust int16

func (e Thrust) Int() int16 {
	return int16(e)
}

func (e Thrust) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Thrust(a)
	})
}

type Mine int16

func (e Mine) Int() int16 {
	return int16(e)
}

func (e Mine) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Mine(a)
	})
}

type Reproduce int16

func (e Reproduce) Int() int16 {
	return int16(e)
}

func (e Reproduce) Run(m *Machine, code []int16) {
	runInstruction(m, code, func() {
		if len(m.stack) <= 0 {
			return
		}
		a := m.stack.Pop()
		m.state.Reproduce(a)
	})
}
