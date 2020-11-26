package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type (
	// Gene represents one gene of the bot's program.
	//
	// A Gene consists of two sections, an evaluation
	// and an execution section.
	// By the end of the evaluation section, the stack
	// will be popped. If the value is > 0 the execution
	// section will be executed.
	Gene struct {
		Evaluate AST
		Execute  AST
	}

	GeneDrawer func(*UI, *ebiten.Image)
)

func NewGene() *Gene {
	return &Gene{
		Evaluate: make([]Instruction, 0),
		Execute:  make([]Instruction, 0),
	}
}

func (d GeneDrawer) DrawCode(ui *UI, img *ebiten.Image) {
	d(ui, img)
}

func GeneDraweFor(i int, g *Gene) GeneDrawer {
	return func(ui *UI, img *ebiten.Image) {
		text.Draw(img,
			fmt.Sprintf(`Gene (%d)
`, i,
			),
			ui.game.assets.font,
			24, 80,
			color.White)
	}
}
