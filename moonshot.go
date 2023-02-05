package main

import (
	"image/png"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/pkg/errors"
	"golang.org/x/image/font"
)

const (
	// windoWidth is the base to determine the total rendered size of the screen
	//
	// The window is then scaled. The height is determined by retrieving the
	// aspect ration of the screen size.
	windowWidth = 1280
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
	// window
	w, h := ebiten.ScreenSizeInFullscreen()
	ratio := float64(w) / float64(h)
	g := &Game{p: &Player{}}
	g.w = windowWidth
	g.h = int(float64(g.w) / ratio)
	ebiten.SetWindowSize(g.w, g.h)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowResizable(false)
	ebiten.SetFullscreen(true)

	g.cyclesPerTick = 1
	g.camera = &camera{
		Position: cp.Vector{X: 0, Y: 0},
		ViewPort: cp.Vector{
			X: float64(g.w),
			Y: float64(g.h),
		},
		zoomStep: 1,
	}
	g.settings.cameraMoveSpeed = 10
	g.settings.inputMap = defaultInputMap()
	g.space = cp.NewSpace()
	g.space.Iterations = 1
	// see https://chipmunk-physics.net/release/ChipmunkLatest-Docs/#cpSpace-SpatialHash
	// Experimenting with the spatial index
	g.space.UseSpatialHash(2.0, 10000)
	g.ui = NewUI(g)

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

	var scenario string
	if len(os.Args) == 2 {
		scenario = os.Args[1]
	}
	if scen, ok := scenarios[scenario]; !ok {
		scenarios["all"].LoadScenario(g)
	} else {
		scen.LoadScenario(g)
	}

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
	ui       *ebiten.Image
	bot      *ebiten.Image
	asteroid *ebiten.Image
	font     font.Face
}

func loadAssets() (*assets, error) {
	a := &assets{}

	fontFile, err := os.Open("gamedata/IBMPlexMono-Regular.ttf")
	if err != nil {
		return nil, errors.Wrap(err, "error loading font")
	}
	defer fontFile.Close()
	fd, err := ioutil.ReadAll(fontFile)
	if err != nil {
		return nil, errors.Wrap(err, "error reading font")
	}
	f, err := truetype.Parse(fd)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing font")
	}
	a.font = truetype.NewFace(f, &truetype.Options{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating font face")
	}

	ui, err := ebitenutil.OpenFile("gamedata/ui.png")
	if err != nil {
		return nil, errors.Wrap(err, "error loading ui sprite")
	}
	defer ui.Close()
	img, err := png.Decode(ui)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding ui sprite")
	}
	a.ui = ebiten.NewImageFromImage(img)

	bot, err := ebitenutil.OpenFile("gamedata/bot.png")
	if err != nil {
		return nil, errors.Wrap(err, "error loading bot sprite")
	}
	defer bot.Close()
	img, err = png.Decode(bot)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding bot sprite")
	}
	a.bot = ebiten.NewImageFromImage(img)

	asteroid, err := ebitenutil.OpenFile("gamedata/asteroid.png")
	if err != nil {
		return nil, errors.Wrap(err, "error loading asteroid sprite")
	}
	defer asteroid.Close()
	img, err = png.Decode(asteroid)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding asteroid sprite")
	}
	a.asteroid = ebiten.NewImageFromImage(img)

	return a, nil
}
