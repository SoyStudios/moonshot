package main

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	os.Exit(runMain())
}

func runMain() int {
	errLog := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	errLog = log.WithPrefix(errLog,
		"t", log.DefaultTimestampUTC,
		"level", "error",
		"caller", log.DefaultCaller,
	)
	ebiten.SetFullscreen(true)
	g := &Game{p: &Player{}}
	g.viewport.scale = 1
	g.viewport.x, g.viewport.y = 0, 0

	if err := ebiten.RunGame(g); err != nil {
		// nolint: errcheck
		errLog.Log("msg", "error during run",
			"err", err,
		)
		return 1
	}
	return 0
}
