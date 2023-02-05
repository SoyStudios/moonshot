package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/jakecoffman/cp"
)

const botFrictionCoeff = 0.4

type (
	Bot struct {
		*cp.Body
		*cp.Shape
		space *cp.Space

		id int16

		// Components

		// leonhardEfficiency (< 1) is the efficiency of the
		// Leonhard Reactor. How efficiently we can convert
		// energy to matter and back.
		//
		// Implemented as a lambda, possible hook for later
		// upgrade mechanics.
		leonhardEfficiency func() float64
		// thrustStep translates a given thrust step value
		// to a force
		thrustStep func(int16) float64
		// scan FOV in degrees
		scanFOV func() float64

		impulses []cp.Vector
		thrust   cp.Vector
		angle    float64

		machine *Machine
	}
)

func BotRunner(bc <-chan *Bot, done chan struct{}) {
	for {
		b, ok := <-bc
		if !ok {
			break
		}
		b.machine.Run()
		done <- struct{}{}
	}
}

func NewBot(sp *cp.Space, id int16) *Bot {
	b := &Bot{
		Body: sp.AddBody(cp.NewBody(100, 10)),

		space: sp,

		id: id,

		leonhardEfficiency: func() float64 {
			return .65
		},
		thrustStep: func(step int16) float64 {
			switch true {
			case step < 100:
				return 80
			case step < 200:
				return 140
			default:
				return 200
			}
		},

		impulses: make([]cp.Vector, 0),
		thrust:   cp.Vector{},

		machine: NewMachine(),
	}
	// connect machine state interface
	b.machine.state = b
	// create shape
	b.Shape = cp.NewCircle(b.Body, 8, cp.Vector{})
	b.Shape.SetElasticity(0)
	b.Shape.SetFriction(0)
	b.Shape.UserData = b
	b.Shape.Filter.Categories = SHAPE_CATEGORY_BOT
	sp.AddShape(b.Shape)

	b.Body.UserData = b

	return b
}

func (b *Bot) Mass() float64 {
	return b.Body.Mass()
}

func (b *Bot) CenterOfGravity() cp.Vector {
	return b.Body.CenterOfGravity()
}

func (b *Bot) FrameReset() {
	b.impulses = b.impulses[:0]
}

func (b *Bot) Reset() {
	b.thrust.X, b.thrust.Y = 0, 0
}

func (b *Bot) X() int16 {
	return int16(math.Round(b.Velocity().X))
}

func (b *Bot) Y() int16 {
	return int16(math.Round(b.Velocity().Y))
}

func (b *Bot) Energy() int16 {
	return int16(math.Round(b.Mass() * b.leonhardEfficiency()))
}

func (b *Bot) ID() int16 {
	return b.id
}

func (b *Bot) RemoteID(int16) int16 {
	return 1
}

func (b *Bot) Scan(x, y int16) (int16, int16) {
	return 0, 0
}

func (b *Bot) Thrust(x, y int16) {
	v := cp.Vector{X: float64(x), Y: float64(y)}
	b.thrust = b.thrust.Add(v)
}

func (b *Bot) Turn(a int16) {
	angle := float64(a) / 180 * math.Pi
	b.angle += angle
}

func (b *Bot) Mine(strength int16) {
}

func (b *Bot) Reproduce(energy int16) {
}

func (b *Bot) Impulse(strength int16) {
	v := cp.ForAngle(b.angle)
	v = v.Mult(float64(strength))
	b.thrust = b.thrust.Add(v)
}

func (b *Bot) Execute() {
	if b.thrust.X != 0 || b.thrust.Y != 0 {
		// apply thrust
		v := b.thrust
		v = v.Clamp(300)
		b.ApplyImpulseAtLocalPoint(
			v,
			b.CenterOfGravity(),
		)
		b.impulses = append(b.impulses, v)
	}
}

const radDeg = 180 / math.Pi

func (b *Bot) DrawInfo(ui *UI, img *ebiten.Image) {
	text.Draw(img,
		fmt.Sprintf(`bot (%d)

  Position: (%.2f, %.2f)
  Heading: %.2f rad (%.2f°)
  Velocity: (%.2f, %.2f)
  Speed: %.2f

  Thrust Vector: (%.2f, %.2f)

  Mass/Energy: %.2f / %.2f
`, b.id,
			b.Position().X, b.Position().Y,
			b.Velocity().ToAngle(), math.Mod((b.Velocity().ToAngle()*radDeg+90), 360),
			b.Velocity().X, b.Velocity().Y,
			b.Velocity().Length(),

			b.thrust.X, b.thrust.Y,

			b.Mass(), b.Mass()*b.leonhardEfficiency(),
		),
		ui.game.assets.font,
		24, 80,
		color.White)

	text.Draw(img,
		"Machine",
		ui.game.assets.font,
		24, 240,
		color.White)
	var buf strings.Builder
	buf.WriteString("  Registers\n\n")
	for i, v := range b.machine.registers {
		buf.WriteString(fmt.Sprintf("  %2d  % 6d\n", i, v))
	}
	text.Draw(img,
		buf.String(),
		ui.game.assets.font,
		24, 260,
		color.White)

	text.Draw(img,
		"Genes",
		ui.game.assets.font,
		200, 260,
		color.White)
	white := ebiten.NewImage(8, 5)
	white.Fill(color.White)
	green := ebiten.NewImage(8, 5)
	green.Fill(color.RGBA{0, 255, 0, 255})
	var row, col int
	op := &ebiten.DrawImageOptions{}
	for i := range b.machine.program {
		if i%10 == 0 {
			row++
			col = 0
		}
		op.GeoM.Reset()
		op.GeoM.Translate(200+(float64(col)*10), 265+(float64(row)*8))
		if b.machine.activated[i] {
			img.DrawImage(green, op)
		} else {
			img.DrawImage(white, op)
		}
		col++
	}
}
