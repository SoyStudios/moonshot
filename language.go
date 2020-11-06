package main

const (
	ILLEGAL Token = 0
	EOF           = 1
	WS            = 2

	CONST = 3

	BGN = 4 // Start gene segment
	END = 5 // End gene segment
	EXE = 6 // Execute gene if pop n == 1

	RDX = 7  // Read X vector and push it on the stack
	RDY = 8  // Read Y vector and push it on the stack
	RDA = 9  // Read ange ands push it on the stack
	RDH = 10 // Read health and push it on the stack
	RDE = 10 // Read total energy and push it on the stack

	PSH = 11 // Push
	POP = 12 // Pop
	JMP = 13 // Jump
	CON = 14 // Constant identifier
	REG = 15 // Register identifier
	REM = 15 // Remote register identifier
	LOC = 16 // Local identifier

	// comparison
	// x COMP y, where x was pushed before y
	GEQ = 17 // Pushes 1 if x >= y, else 0
	SEQ = 18 // Pushes 1 if x <= y, else 0
	IEQ = 19 // Pushes 1 if x == y, else 0

	THR = 20 // Pop and thrust for n units
	TRN = 21 // Turn by n degrees
)

type (
	Token int16

	Instruction interface {
		Int() int16
	}
)

func Translate(token Token) Instruction {
	switch token {
	case BGN:
		return Begin(BGN)
	case END:
		return End(END)
	case EXE:
		return Exec(EXE)
	case ILLEGAL:
		fallthrough
	default:
		return Illegal(ILLEGAL)
	}
}

func Program(tks []Token) []int16 {
	program := make([]int16, len(tks))
	for i, t := range tks {
		program[i] = Translate(t).Int()
	}
	return program
}

// instructions
type Illegal int16

func (i Illegal) Int() int16 {
	return int16(i)
}

func (i Illegal) Run(m *Machine) {}

type Begin int16

func (b Begin) Int() int16 {
	return int16(b)
}

type End int16

func (e End) Int() int16 {
	return int16(e)
}

type Exec int16

func (e Exec) Int() int16 {
	return int16(e)
}

func (e Exec) Run(m *Machine) {
	if m.pc > len(m.program)+1 {
		return
	}
	inst := Translate(Token(m.program[m.pc]))
	m.pc++
	_ = inst
}
