package main

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-kit/log"
	"github.com/jakecoffman/cp"
	"github.com/pkg/errors"
)

const (
	// windoWidth is the base to determine the total rendered size of the screen
	//
	// The window is then scaled. The height is determined by retrieving the
	// aspect ration of the screen size.
	windowWidth = 1280

	title = "moonshot"
)

var ErrExit = errors.New("exit")

func main() {
	os.Exit(runMain())
}

var errLog, infoLog log.Logger

func runMain() int {
	errLog = log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	errLog = log.WithPrefix(errLog,
		"t", log.DefaultTimestampUTC,
		"level", "error",
		"caller", log.DefaultCaller,
	)
	infoLog = log.NewSyncLogger(log.NewLogfmtLogger(os.Stdout))
	infoLog = log.WithPrefix(infoLog,
		"t", log.DefaultTimestampUTC,
		"level", "info",
		"caller", log.DefaultCaller,
	)

	// window
	w, h := rl.GetScreenWidth(), rl.GetScreenHeight()
	ratio := float64(w) / float64(h)
	rl.InitWindow(windowWidth, int32(float64(w)/ratio), title)
	g := &Game{p: &Player{}}
	g.w = windowWidth
	g.h = int(float64(g.w) / ratio)

	g.cyclesPerTick = 1
	g.space = cp.NewSpace()
	g.space.Iterations = 1
	// see https://chipmunk-physics.net/release/ChipmunkLatest-Docs/#cpSpace-SpatialHash
	// Experimenting with the spatial index
	g.space.UseSpatialHash(2.0, 10000)

	g.init()

	var scenario string
	if len(os.Args) == 2 {
		scenario = os.Args[1]
	}
	if scen, ok := scenarios[scenario]; !ok {
		scenarios["all"].LoadScenario(g)
	} else {
		scen.LoadScenario(g)
	}

	return 0
}
