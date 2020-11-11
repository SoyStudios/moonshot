package main

import (
	"image/png"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pkg/errors"
	"golang.org/x/image/math/f64"
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
	w, h := ebiten.WindowSize()
	g := &Game{p: &Player{}}
	g.camera = &camera{
		ViewPort: f64.Vec2{
			float64(w),
			float64(h),
		},
		zoomFactor: 1,
	}

	var err error
	g.assets, err = loadAssets()
	if err != nil {
		// nolint: errcheck
		errLog.Log("msg", "error during load",
			"err", err,
		)
		return 1
	}

	if err := ebiten.RunGame(g); err != nil {
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
