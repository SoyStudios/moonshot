package main

import (
	"fmt"
	"image/color"
	"math"
	"runtime"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
)

const baseZoomFactor = 1.01

type (
	camera struct {
		ViewPort cp.Vector
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

		cyclesPerTick int

		space *cp.Space
		world *ebiten.Image

		bots []*Bot

		numRunners int
		botChan    chan *Bot
		doneChan   chan struct{}

		ui *UI

		w, h int

		p *Player
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

func (c *camera) worldMatrix() ebiten.GeoM {
	m := ebiten.GeoM{}
	// We want to scale and rotate around center of image / screen
	m.Translate(-c.viewportCenter().X, -c.viewportCenter().Y)
	m.Rotate(float64(c.rotation) * 2 * math.Pi / 360)
	m.Translate(c.viewportCenter().X, c.viewportCenter().Y)
	return m
}

// worldObjectMatrix returns a matrix used to place an object
// onto the world on coordinates x, y
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
	screen.DrawImage(world, &ebiten.DrawImageOptions{
		GeoM: c.worldMatrix(),
	})
}

// WorldToScreen translates world coordinates (such as positions of bots
// etc.) into screen coordinates for rendering onto the world plane
func (c *camera) WorldToScreen(x, y float64) (float64, float64) {
	vec := cp.Vector{X: x, Y: y}
	vec = vec.Add(c.Position.Neg())
	return vec.X, vec.Y
}

func (c *camera) ScreenToWorld(posX, posY int) (float64, float64) {
	inverseMatrix := c.worldMatrix()
	inverseMatrix.Translate(-c.Position.X, -c.Position.Y)
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
	g.world = ebiten.NewImage(g.w, g.h)

	b := NewBot(g.space, 1)
	b.SetPosition(cp.Vector{X: 0, Y: 100})
	b.SetVelocity(100, 0)

	g.bots = []*Bot{b}

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
	RDX
	RDY
	ABS
	PSH CON 100
	LST
END
BEGIN EX
	PSH CON 50
	PSH CON 50
	THR
	PSH CON 1
	POP REG 0
END

BEGIN EV
	PSH REG 0
END
BEGIN EX
	PSH CON 50
	NEG
	PSH CON 50
	THR
END

BEGIN EV
	RDX
	RDY
	ABS
	PSH CON 150
	GEQ
END
BEGIN EX
	PSH CON 0
	POP REG 0
	PSH CON 1
	POP REG 1 
END

BEGIN EV
	PSH REG 1
END
BEGIN EX
	PSH CON 50
	NEG
	PSH CON 50
	NEG
	THR
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

BEGIN EV
	// if reg2 < 1
	PSH REG 2
	PSH CON 1
	LST
END
BEGIN EX
	PSH CON 1
	POP REG 2
	PSH CON 10
	POP REG 3
END

BEGIN EV
	// if reg2 >= 1
	PSH REG 2
	PSH CON 1
	GEQ
END
BEGIN EX
	PSH REG 3
	TRN
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
	g.ui.info = b
	g.ui.code = GeneDrawerFor(0, b.machine.program[0])

	g.numRunners = runtime.NumCPU() - 1
	if g.numRunners < 2 {
		g.numRunners = 2
	}
	g.doneChan = make(chan struct{}, 1)
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

// Update is the main update loop
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrExit
	}

	g.ui.Update()

	// Camera controls
	g.updateOnKey("zoomOut", func() {
		g.camera.zoomStep--
		g.camera.zoomTo(g.camera.ScreenToWorld(ebiten.CursorPosition()))
	})
	g.updateOnKey("zoomIn", func() {
		g.camera.zoomStep++
		g.camera.zoomTo(g.camera.ScreenToWorld(ebiten.CursorPosition()))
	})
	g.updateOnKey("up", func() {
		g.camera.Position.Y -= g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})
	g.updateOnKey("down", func() {
		g.camera.Position.Y += g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})
	g.updateOnKey("left", func() {
		g.camera.Position.X -= g.settings.cameraMoveSpeed / g.camera.zoomFactor()
	})
	g.updateOnKey("right", func() {
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
		g.botChan = make(chan *Bot, 1)
		for i := 0; i < g.numRunners; i++ {
			go BotRunner(g.botChan, g.doneChan)
		}
		for _, bot := range g.bots {
			g.botChan <- bot
		}
		close(g.botChan)
		for i := 0; i < len(g.bots); i++ {
			<-g.doneChan
		}
		tps := ebiten.CurrentTPS()
		if tps == 0 {
			tps = 60
		}
		g.space.Step(1.0 / tps)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	botSizeX, botSizeY := g.assets.bot.Size()
	botDX, botDY := float64(botSizeX)/2, float64(botSizeY)/2
	for _, bot := range g.bots {
		op.GeoM = g.camera.worldObjectMatrix(
			bot.Position().X-botDX,
			bot.Position().Y-botDY,
		)
		g.world.DrawImage(g.assets.bot, op)

		// draw viewing angle
		// start position matrix
		ms := g.camera.worldObjectMatrix(0, 0)
		dir := cp.ForAngle(bot.angle)
		dir = dir.Clamp(1)
		dir = dir.Mult(32)
		me := g.camera.worldObjectMatrix(dir.X, dir.Y)
		sx, sy := ms.Apply(bot.Position().X, bot.Position().Y)
		dx, dy := me.Apply(bot.Position().X, bot.Position().Y)
		ebitenutil.DrawLine(g.world,
			sx, sy,
			dx, dy,
			color.RGBA{255, 0, 0, 255},
		)

		// draw impulses
		for _, imp := range bot.impulses {
			ms = g.camera.worldObjectMatrix(bot.CenterOfGravity().X, bot.CenterOfGravity().Y)
			sx, sy = ms.Apply(bot.Position().X, bot.Position().Y)
			me = g.camera.worldObjectMatrix(imp.X, imp.Y)
			dx, dy = me.Apply(bot.Position().X, bot.Position().Y)
			ebitenutil.DrawLine(g.world,
				sx, sy,
				dx, dy,
				color.RGBA{0, 255, 0, 255},
			)
		}

		bot.FrameReset()
	}

	g.camera.Render(g.world, screen)

	worldX, worldY := g.camera.ScreenToWorld(ebiten.CursorPosition())
	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf("TPS: %0.2f, C: %d\nB: %.2f,%.2f, M: %.2f, BA: %.2f, BAV: %.2f",
			ebiten.CurrentTPS(), g.cyclesPerTick,
			g.bots[2].Velocity().X, g.bots[2].Velocity().Y,
			math.Sqrt(math.Pow(g.bots[2].Velocity().X, 2)+math.Pow(g.bots[2].Velocity().Y, 2)),
			180/math.Pi*g.bots[2].Angle(), g.bots[2].AngularVelocity(),
		),
	)

	g.ui.Draw(screen)

	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("%s\nCursor World Pos: %.2f,%.2f",
			g.camera.String(),
			worldX, worldY,
		),
		0, g.h-48,
	)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return g.w, g.h
}
