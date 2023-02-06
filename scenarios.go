package main

import (
	"image"
	"strings"
	"time"

	"github.com/jakecoffman/cp"
)

type (
	Scenario interface {
		LoadScenario(g *Game)
	}

	ScenarioFunc func(g *Game)
)

func (f ScenarioFunc) LoadScenario(g *Game) {
	f(g)
}

var scenarios = map[string]ScenarioFunc{
	"all": func(g *Game) {

		b := NewBot(g.space, 1)
		b.SetPosition(cp.Vector{X: 0, Y: 100})
		b.SetVelocity(100, 0)
		g.bots = append(g.bots, b)

		b = NewBot(g.space, 1)
		b.SetPosition(cp.Vector{X: 600, Y: 100})
		b.SetVelocity(-10, 0)
		g.bots = append(g.bots, b)

		code := `
BEGIN EV
	// Read bot’s current energy level and push it to the stack
	RDE
	PSH CON 1000
	GEQ
END
BEGIN EX
	PSH CON 500
	REP
END

BEGIN EV
	// If total velocity is >= 200
	RDX
	RDY
	ABS
	PSH CON 200
	GEQ
END
BEGIN EX
	PSH CON 0
	POP REG 0
	// thrust in opposite direction
	RDX
	NEG
	RDY
	NEG
	THR
END

// Counter for turning
// Register 0 holds counter
BEGIN EV
	// If reg0 <= 80
	PSH REG 0
	PSH CON 80
	LEQ
END
BEGIN EX
	// reg0++
	PSH REG 0
	PSH CON 1
	ADD
	POP REG 0
END

// Turning every 80 ticks
BEGIN EV
	// if reg0 > 80
	PSH REG 0
	PSH CON 80
	GRT
END
BEGIN EX
	// reset reg0 to 0
	// turn by 10 degrees
	PSH CON 0
	POP REG 0
	PSH CON 10
	TRN
	PSH CON 500
	IMP
END

// Create impulse every 20 ticks
// counter in reg1
BEGIN EV
	// if reg1 <= 20
	PSH REG 1
	PSH CON 20
	LEQ
END
BEGIN EX
	// reg1++
	PSH REG 1
	PSH CON 1
	ADD
	POP REG 1
END

// Create impulse in current direction
BEGIN EV
	// if reg1 > 20
	PSH REG 1
	PSH CON 20
	GRT
END
BEGIN EX
	PSH CON 500
	IMP
	PSH CON 0
	POP REG 1
END
	`
		p := NewParser(strings.NewReader(code))
		program, err := p.Parse()
		if err != nil {
			panic(err)
		}

		b = NewBot(g.space, 1)
		b.SetPosition(cp.Vector{X: 200, Y: 200})
		b.machine.program = program
		g.bots = append(g.bots, b)
		g.controls.follow = b

		g.ui.info = b
	},

	"asteroid": func(g *Game) {
		a := NewAsteroid(image.Rect(0, 0, 1000, 1000))
		a.generate(time.Now().Unix())
		g.asteroids = append(g.asteroids, a)
	},
}
