// Command crochet is the demo front-end for the crochet engine. It builds a
// handful of crochet pieces and simulates their yarn physics in a 3-D window.
//
//	go run ./crochet/cmd/crochet
//
// Controls: drag to orbit, wheel to zoom, right-drag to pan, [tab] next scene,
// [space] pause, [r] reset, [w] wind, [l]/[n] toggle links/nodes.
package main

import (
	"os"
	"strconv"

	"github.com/SoyStudios/moonshot/crochet/engine"
	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/pattern"
	"github.com/SoyStudios/moonshot/crochet/physics"
	"github.com/SoyStudios/moonshot/crochet/yarn"
)

func main() {
	cfg := engine.DefaultConfig()
	cfg.Title = "Crochet Engine — yarn physics"

	builders := []func() *engine.Scene{
		hangingSwatch,
		drapedSwatch,
		stuffedBall,
		stripedBeanie,
		materialShowcase,
		amigurumiDisc,
	}

	// Optional headless preview: CROCHET_SHOT=out.png [CROCHET_SCENE=n]
	// [CROCHET_FRAMES=n] renders one scene, saves a PNG and exits.
	if shot := os.Getenv("CROCHET_SHOT"); shot != "" {
		cfg.Screenshot = shot
		cfg.ScreenshotFrames = envInt("CROCHET_FRAMES", 300)
		if n := envInt("CROCHET_SCENE", 0); n >= 0 && n < len(builders) {
			builders = builders[n : n+1]
		}
	}

	engine.New(cfg, builders...).Run()
}

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

// wool returns a soft, mostly-matte yarn in the given colour.
func wool(r, g, b uint8) yarn.Config {
	c := yarn.DefaultConfig()
	c.Color = yarn.Color{R: r, G: g, B: b, A: 255}
	c.Material = yarn.Material{Sheen: 0.12, Ambient: 0.36}
	return c
}

// silk returns a glossier yarn with a stronger specular sheen.
func silk(r, g, b uint8) yarn.Config {
	c := wool(r, g, b)
	c.Material = yarn.Material{Sheen: 0.75, Ambient: 0.30}
	return c
}

// hangingSwatch: a rectangular sampler pinned along its top edge, hanging and
// swaying under gravity — the "hello world" of the yarn simulation.
func hangingSwatch() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true

	f := pattern.Swatch(w, pattern.SwatchConfig{
		Rows:   16,
		Cols:   22,
		Origin: math3.V(-5.25, 9, 0),
		U:      math3.V(0.5, 0, 0),
		V:      math3.V(0, -0.5, 0),
		Stitch: pattern.Single,
		Pin:    pattern.PinTopEdge,
		Yarn:   wool(214, 118, 74),
	})

	return &engine.Scene{Name: "Hanging swatch (sc, top-edge pinned)", World: w, Fabrics: []*pattern.Fabric{f}}
}

// drapedSwatch: a horizontal sheet of crochet dropped over a sphere so the
// fabric folds and drapes around the form, with self-collision so the folds
// don't pass through each other.
func drapedSwatch() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true
	w.SelfCollision = true
	w.CollisionRadius = 0.2
	w.Spheres = []physics.Sphere{{Center: math3.V(0, 2.4, 0), Radius: 2.4}}

	f := pattern.Swatch(w, pattern.SwatchConfig{
		Rows:   26,
		Cols:   26,
		Origin: math3.V(-6.5, 6.2, -6.5),
		U:      math3.V(0.5, 0, 0),
		V:      math3.V(0, 0, 0.5),
		Stitch: pattern.HalfDouble,
		Pin:    pattern.PinNone,
		Yarn:   wool(120, 170, 210),
	})

	return &engine.Scene{Name: "Draped swatch over a form (hdc, self-collision)", World: w, Fabrics: []*pattern.Fabric{f}}
}

// stuffedBall: an amigurumi ball built from per-round stitch counts, sealed at
// both poles, inflated by a stuffing (pressure) constraint and held together
// by self-collision. It settles on the ground like a stuffed toy.
func stuffedBall() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true
	w.GroundY = 0
	w.SelfCollision = true
	w.CollisionRadius = 0.16

	f := pattern.Revolve(w, pattern.RevolveConfig{
		Counts:      pattern.SphereCounts(6, 6), // 6,12,…,36,…,12,6
		Stitch:      pattern.Single,
		Gauge:       0.55,
		Center:      math3.V(0, 3.2, 0),
		CloseBottom: true,
		CloseTop:    true,
		Yarn:        wool(230, 120, 140),
	})
	f.Stuff(0.006) // internal stuffing pressure

	return &engine.Scene{Name: "Stuffed amigurumi ball (pressure + self-collision)", World: w, Fabrics: []*pattern.Fabric{f}}
}

// stripedBeanie: a hat worked in the round with a self-striping colourway.
func stripedBeanie() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true

	y := wool(240, 240, 240)
	y.Stripe = []yarn.Color{
		{R: 214, G: 69, B: 65, A: 255},
		{R: 244, G: 180, B: 60, A: 255},
		{R: 60, G: 150, B: 200, A: 255},
	}
	y.StripeWidth = 2 // rounds per colour band

	f := pattern.Tube(w, pattern.TubeConfig{
		Rounds:      20,
		Stitches:    28,
		Radius:      2.2,
		RiseY:       0.42,
		Center:      math3.V(0, 1.0, 0),
		PinTopRound: true,
		Yarn:        y,
	})

	return &engine.Scene{Name: "Striped beanie (self-striping colourwork)", World: w, Fabrics: []*pattern.Fabric{f}}
}

// materialShowcase: two identical hanging swatches side by side, one matte
// wool, one glossy silk, to show the material lighting difference.
func materialShowcase() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true

	mk := func(originX float64, y yarn.Config) *pattern.Fabric {
		return pattern.Swatch(w, pattern.SwatchConfig{
			Rows:   14, Cols: 12,
			Origin: math3.V(originX, 9, 0),
			U:      math3.V(0.5, 0, 0),
			V:      math3.V(0, -0.5, 0),
			Stitch: pattern.Double,
			Pin:    pattern.PinTopEdge,
			Yarn:   y,
		})
	}
	matte := mk(-6.5, wool(190, 90, 110))
	glossy := mk(0.8, silk(190, 90, 110))

	return &engine.Scene{Name: "Materials: matte wool vs glossy silk (dc)", World: w, Fabrics: []*pattern.Fabric{matte, glossy}}
}

// amigurumiDisc: a flat increasing circle (the amigurumi base) pinned at its
// centre so it relaxes into a gently rippling disc.
func amigurumiDisc() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true

	f := pattern.Disc(w, pattern.DiscConfig{
		Rounds:      9,
		StartStitch: 6,
		Increase:    6,
		RingSpacing: 0.7,
		Center:      math3.V(0, 6, 0),
		PinCenter:   true,
		Yarn:        wool(230, 160, 90),
	})

	return &engine.Scene{Name: "Amigurumi disc (magic ring, 6-st increases)", World: w, Fabrics: []*pattern.Fabric{f}}
}
