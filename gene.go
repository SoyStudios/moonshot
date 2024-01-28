package main

import (
	"strings"
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
)

func NewGene() *Gene {
	return &Gene{
		Evaluate: make([]Instruction, 0),
		Execute:  make([]Instruction, 0),
	}
}

func (g *Gene) String() string {
	var b strings.Builder
	b.WriteString(g.Evaluate.String())
	b.WriteString(g.Execute.String())
	return b.String()
}
