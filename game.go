package main

import (
	"fmt"
	"image/color"
	"math"
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

// The camera has zoom steps, the factor by which each step is
// multiplied
const baseZoomFactor = 1.01

type (
	camera struct {
		// ViewPort is the size of the viewport width * height
		ViewPort cp.Vector
		// Position of the camera in the world
		Position cp.Vector
		zoomStep int
		rotation int
	}

	Game struct {
		settings struct {
			cameraMoveSpeed float64

			inputMap map[string][]ebiten.Key
		}

		paused bool
		step   int

		assets *assets

		camera *camera

		controls struct {
			follow Positioner
		}

		cyclesPerTick int

		space *cp.Space
		world *ebiten.Image

		bots []*Bot

		numRunners int
		wg         sync.WaitGroup
		botChan    chan *Bot

		ui *UI

		// width and height of the game scene in pixels
		w, h int

		p *Player
	}

	Positioner interface {
		Position() cp.Vector
	}
)

func (c *camera) String() string {
	return fmt.Sprintf(
		"T: %.1f, VP: %.1f, R: %d, S: %d, Z: %.1f",
		c.Position, c.ViewPort, c.rotation, c.zoomStep, c.zoomFactor(),
	)
}

func (c *camera) viewportCenter() cp.Vector {
	return cp.Vector{
		X: c.ViewPort.X * 0.5,
		Y: c.ViewPort.Y * 0.5,
	}
}

// worldObjectMatrix returns a matrix used to place an object
// onto the world on coordinates x, y
// relative to the camera
func (c *camera) worldObjectMatrix(x, y float64) ebiten.GeoM {
	g := ebiten.GeoM{}
	g.Translate(-c.Position.X, -c.Position.Y)
	g.Translate(x, y)
	g.Translate(-c.viewportCenter().X, -c.viewportCenter().Y)
	g.Scale(c.zoomFactor(), c.zoomFactor())
	g.Translate(c.viewportCenter().X, c.viewportCenter().Y)
	return g
}

func (c *camera) Render(world, screen *ebiten.Image) {
	screen.DrawImage(world, &ebiten.DrawImageOptions{})
}

// WorldToScreen translates world coordinates (such as positions of bots
// etc.) into screen coordinates for rendering onto the world plane
func (c *camera) WorldToScreen(x, y float64) (float64, float64) {
	vec := cp.Vector{X: x, Y: y}
	vec = vec.Add(c.Position.Neg())
	return vec.X, vec.Y
}

func (c *camera) ScreenToWorld(posX, posY int) (float64, float64) {
	inverseMatrix := c.worldObjectMatrix(0, 0)
	if inverseMatrix.IsInvertible() {
		inverseMatrix.Invert()
		return inverseMatrix.Apply(float64(posX), float64(posY))
	} else {
		// When scaling it can happend that matrix is not invertable
		return math.NaN(), math.NaN()
	}
}

func (c *camera) zoomFactor() float64 {
	return math.Pow(baseZoomFactor, float64(c.zoomStep))
}

func (c *camera) zoomTo(x, y float64) {
	to := cp.Vector{X: x, Y: y}
	to = to.Clamp(1)
	to = to.Mult(1 / c.zoomFactor())
	c.Position = c.Position.Add(to)
}

func (g *Game) init() {
	initBotDraw(g)

	g.paused = true
	g.world = ebiten.NewImage(g.w, g.h)
	g.bots = make([]*Bot, 0, 128)

	g.numRunners = runtime.NumCPU() - 1
	if g.numRunners < 2 {
		g.numRunners = 2
	}
	g.botChan = make(chan *Bot, 1)
	for i := 0; i < g.numRunners; i++ {
		go BotRunner(g, g.botChan)
	}
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

	g.ui.Update()

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
			g.step++
		}
	})
	if g.paused && g.step == 0 {
		return nil
	}
	g.step = 0

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
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Fill(color.Black)

	for _, bot := range g.bots {
		bot.Draw(g)
	}

	g.camera.Render(g.world, screen)

	g.ui.Draw(screen)

	// debug info
	worldX, worldY := g.camera.ScreenToWorld(ebiten.CursorPosition())
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("TPS: %0.2f, C: %d\n%s\nCursor World Pos: %.2f,%.2f\n%s",
			ebiten.CurrentTPS(), g.cyclesPerTick,
			g.camera.String(),
			worldX, worldY,
			func() string {
				if g.paused {
					return "*PAUSED*"
				}
				return ""
			}(),
		),
		0, g.h-72,
	)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return g.w, g.h
}
