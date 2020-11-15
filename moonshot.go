package main

import (
	"image/png"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/pkg/errors"
	"golang.org/x/image/math/f64"
)

const (
	windowWidth = 1024
)

var ErrExit = errors.New("exit")

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
	w, h := ebiten.ScreenSizeInFullscreen()
	ratio := float64(w) / float64(h)
	g := &Game{p: &Player{}}
	g.w = windowWidth
	g.h = int(float64(g.w) / ratio)
	ebiten.SetWindowSize(g.w, g.h)
	ebiten.SetFullscreen(true)
	g.tps = 60
	g.camera = &camera{
		Position: f64.Vec2{100, 100},
		ViewPort: f64.Vec2{
			float64(g.w),
			float64(g.h),
		},
		zoomFactor: 1,
	}
	g.space = cp.NewSpace()
	g.space.Iterations = 1
	g.space.UseSpatialHash(2.0, 10000)

	var err error
	g.assets, err = loadAssets()
	if err != nil {
		// nolint: errcheck
		errLog.Log("msg", "error during load",
			"err", err,
		)
		return 1
	}

	g.init()

	if err := ebiten.RunGame(g); err != nil {
		if err == ErrExit {
			return 0
		}
		// nolint: errcheck
		errLog.Log("msg", "error during run",
			"err", err,
		)
		return 1
	}
	return 0
}

type assets struct {
	bot *ebiten.Image
}

func loadAssets() (*assets, error) {
	bot, err := ebitenutil.OpenFile("gamedata/bot.png")
	if err != nil {
		return nil, errors.Wrap(err, "error loading bot sprite")
	}
	defer bot.Close()
	img, err := png.Decode(bot)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding bot sprite")
	}

	a := &assets{}
	a.bot = ebiten.NewImageFromImage(img)

	return a, nil
}
