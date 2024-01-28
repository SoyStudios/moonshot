package main

import (
	"runtime"
	"sync"

	"github.com/jakecoffman/cp"
)

// The shape categories for chipmunk
// see https://chipmunk-physics.net/release/ChipmunkLatest-Docs/#cpShape-Filtering
const (
	SHAPE_CATEGORY_ANY = 1 << iota
	SHAPE_CATEGORY_BOT
	SHAPE_CATEGORY_ASTEROID
)

type (
	Game struct {
		paused bool
		doStep bool

		cyclesPerTick int
		step          int64

		space *cp.Space

		bots []*Bot

		numRunners int
		wg         sync.WaitGroup
		botChan    chan *Bot

		asteroids []*Asteroid

		// width and height of the game scene in pixels
		w, h int

		p *Player
	}
)

func (g *Game) init() {
	g.paused = true
	g.bots = make([]*Bot, 0, 128)
	g.asteroids = make([]*Asteroid, 0, 64)

	g.numRunners = runtime.NumCPU() - 1
	if g.numRunners < 2 {
		g.numRunners = 2
	}
	g.botChan = make(chan *Bot, 1)
	for i := 0; i < g.numRunners; i++ {
		go BotRunner(g, g.botChan)
	}
}

// Update is the main update loop
func (g *Game) Update() error {
	if g.paused && !g.doStep {
		return nil
	}
	g.doStep = false

	// Run bot cycles
	for i := 0; i < g.cyclesPerTick; i++ {
		g.wg.Add(len(g.bots))
		for _, bot := range g.bots {
			g.botChan <- bot
		}
		g.wg.Wait()
		//g.space.Step(1.0 / float64(ebiten.TPS()))
		g.step++
	}
	return nil
}
