package main

import (
	"github.com/jakecoffman/cp"
)

type (
	Bot struct {
		*cp.Body

		machine *Machine
	}
)
