package main

import (
	"fmt"
	"image/color"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
		settings struct {
			cameraMoveSpeed float64

			inputMap map[string][]ebiten.Key
		}

		paused bool
		doStep bool

		assets *assets

		camera *camera

		controls struct {
			follow Positioner
		}

		cyclesPerTick int
		step          int64

		space *cp.Space
		world *ebiten.Image

		bots []*Bot

		numRunners int
		wg         sync.WaitGroup
		botChan    chan *Bot

		asteroids []*Asteroid

		ui *UI

		physicsDebug *PhyicsDebug

		// width and height of the game scene in pixels
		w, h int

		p *Player
	}

	Positioner interface {
		Position() cp.Vector
	}
)

func (g *Game) init() {
	initBotDraw(g)

	g.paused = true
	g.world = ebiten.NewImage(g.w, g.h)
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

	g.physicsDebug = &PhyicsDebug{g: g}
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (g *Game) updateOnKey(input string, f func()) {
	for _, k := range g.settings.inputMap[input] {
		if ebiten.IsKeyPressed(k) {
			f()
		}
	}
}

func (g *Game) updateOnRepeatingKey(input string, f func()) {
	for _, k := range g.settings.inputMap[input] {
		if repeatingKeyPressed(k) {
			f()
		}
	}
}

func (g *Game) resetFollow() {
	g.controls.follow = nil
}

// Update is the main update loop
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrExit
	}

	// Camera controls
	g.updateOnKey("zoomOut", func() {
		g.resetFollow()
		g.camera.zoomStep--
		g.camera.zoomTo(g.camera.ScreenToWorld(ebiten.CursorPosition()))
	})
	g.updateOnKey("zoomIn", func() {
		g.resetFollow()
		g.camera.zoomStep++
		g.camera.zoomTo(g.camera.ScreenToWorld(ebiten.CursorPosition()))
	})
	g.updateOnKey("up", func() {
		g.resetFollow()
		g.camera.Position.Y -= g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})
	g.updateOnKey("down", func() {
		g.resetFollow()
		g.camera.Position.Y += g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})
	g.updateOnKey("left", func() {
		g.resetFollow()
		g.camera.Position.X -= g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})
	g.updateOnKey("right", func() {
		g.resetFollow()
		g.camera.Position.X += g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})

	// pause
	g.updateOnRepeatingKey("pause", func() {
		g.paused = !g.paused
	})
	g.updateOnRepeatingKey("step", func() {
		if g.paused {
			g.doStep = true
		}
	})

	// debug
	g.updateOnRepeatingKey("physicsDebug", func() {
		g.physicsDebug.enabled = !g.physicsDebug.enabled
	})

	g.ui.Update()

	if g.paused && !g.doStep {
		return nil
	}
	g.doStep = false

	// follow
	if g.controls.follow != nil {
		g.camera.Position.X, g.camera.Position.Y = g.controls.follow.Position().X-g.camera.viewportCenter().X, g.controls.follow.Position().Y-g.camera.viewportCenter().Y
	}

	// Game speed controls
	g.updateOnRepeatingKey("speedUp", func() {
		if ebiten.CurrentTPS() > 10 {
			g.cyclesPerTick *= 2
		}
	})
	g.updateOnRepeatingKey("speedDown", func() {
		g.cyclesPerTick /= 2
		if g.cyclesPerTick <= 0 {
			g.cyclesPerTick = 1
		}
	})

	// Run bot cycles
	for i := 0; i < g.cyclesPerTick; i++ {
		g.wg.Add(len(g.bots))
		for _, bot := range g.bots {
			g.botChan <- bot
		}
		g.wg.Wait()
		g.space.Step(1.0 / float64(ebiten.TPS()))
		g.step++
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Fill(color.Black)

	for _, bot := range g.bots {
		bot.Draw(g)
	}

	for _, asteroid := range g.asteroids {
		asteroid.Draw(g)
	}

	g.camera.Render(g.world, screen)

	if g.physicsDebug.enabled {
		g.physicsDebug.initialize()
		cp.DrawSpace(g.space, g.physicsDebug)
		g.physicsDebug.Draw(screen)
	}

	g.ui.Draw(screen)

	// debug info
	cursorX, cursorY := ebiten.CursorPosition()
	worldX, worldY := g.camera.ScreenToWorld(cursorX, cursorY)
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("TPS: %0.2f, C: %d\n%s\nCursor: %d,%d World Pos: %.2f,%.2f\nStep: %d\n%s %s",
			ebiten.CurrentTPS(), g.cyclesPerTick,
			g.camera.String(),
			cursorX, cursorY,
			worldX, worldY,
			g.step,
			func() string {
				if g.paused {
					return "*PAUSED*"
				}
				return ""
			}(),
			func() string {
				if g.physicsDebug.enabled {
					return "*DEBUG*"
				}
				return ""
			}(),
		),
		0, g.h-128,
	)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return g.w, g.h
}
