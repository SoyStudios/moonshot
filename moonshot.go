package main

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-kit/log"
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
	g := &Game{}
	g.w = windowWidth
	g.h = int(float64(g.w) / ratio)

	rl.InitWindow(int32(g.w), int32(g.h), title)
	defer rl.CloseWindow()

	g.cyclesPerTick = 1

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

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		if rl.IsKeyPressed(rl.KeySpace) {
			g.paused = !g.paused
		}
		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			delta := rl.GetMouseDelta()
			delta = rl.Vector2Scale(delta, -1/g.camera.Zoom)
			g.camera.Target = rl.Vector2Add(g.camera.Target, delta)
		}

		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), g.camera)
			g.camera.Offset = rl.GetMousePosition()
			g.camera.Target = mouseWorldPos
			const zoomIncrement = 0.0125
			zoom := zoomIncrement * wheel
			g.camera.Zoom += zoom
			if g.camera.Zoom <= 0 {
				g.camera.Zoom = zoomIncrement
			}
		}

		g.Update(dt)

		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		rl.BeginMode2D(g.camera)
		for _, b := range g.bots {
			rl.DrawCircleV(
				rl.Vector2{X: float32(b.Position().X), Y: float32(b.Position().Y)},
				8, rl.White)
		}

		//		for _, a := range g.asteroids {
		//
		//		}
		rl.EndMode2D()

		rl.EndDrawing()
	}

	return 0
}
