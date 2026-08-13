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
		beanie,
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

// wool returns a yarn look/feel with the given colour.
func wool(r, g, b uint8) yarn.Config {
	c := yarn.DefaultConfig()
	c.Color = yarn.Color{R: r, G: g, B: b, A: 255}
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
		U:      math3.V(0.5, 0, 0),   // columns run along +X
		V:      math3.V(0, -0.5, 0),  // rows descend along -Y
		Stitch: pattern.Single,
		Pin:    pattern.PinTopEdge,
		Yarn:   wool(214, 118, 74),
	})

	return &engine.Scene{Name: "Hanging swatch (sc, top-edge pinned)", World: w, Fabrics: []*pattern.Fabric{f}}
}

// drapedSwatch: a horizontal sheet of crochet dropped over a sphere so the
// fabric folds and drapes around the form (fabric-over-armature).
func drapedSwatch() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true
	w.Spheres = []physics.Sphere{{Center: math3.V(0, 2.4, 0), Radius: 2.4}}

	f := pattern.Swatch(w, pattern.SwatchConfig{
		Rows:   26,
		Cols:   26,
		Origin: math3.V(-6.5, 6.2, -6.5),
		U:      math3.V(0.5, 0, 0), // columns run along +X
		V:      math3.V(0, 0, 0.5), // rows run along +Z (a flat sheet up high)
		Stitch: pattern.HalfDouble,
		Pin:    pattern.PinNone,
		Yarn:   wool(120, 170, 210),
	})

	return &engine.Scene{Name: "Draped swatch over a form (hdc)", World: w, Fabrics: []*pattern.Fabric{f}}
}

// beanie: a tube worked in the round, pinned at the top round like a hat hung
// from its crown.
func beanie() *engine.Scene {
	w := physics.NewWorld()
	w.Ground = true

	f := pattern.Tube(w, pattern.TubeConfig{
		Rounds:      20,
		Stitches:    28,
		Radius:      2.2,
		RiseY:       0.42,
		Center:      math3.V(0, 1.0, 0),
		PinTopRound: true,
		Yarn:        wool(150, 200, 120),
	})

	return &engine.Scene{Name: "Beanie tube (worked in the round, crown pinned)", World: w, Fabrics: []*pattern.Fabric{f}}
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
