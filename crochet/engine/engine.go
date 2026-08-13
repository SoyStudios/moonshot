// Package engine wires the headless crochet simulation to a raylib window: a
// 3-D orbit camera, tube rendering of the yarn, a fixed-timestep physics loop
// and keyboard/mouse controls. It is the only package that depends on raylib
// (and therefore cgo); everything below it (math3, physics, yarn, pattern) is
// pure Go and unit-tested without a display.
package engine

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SoyStudios/moonshot/crochet/math3"
	"github.com/SoyStudios/moonshot/crochet/pattern"
	"github.com/SoyStudios/moonshot/crochet/physics"
)

// Scene bundles a physics world with the crochet fabrics rendered from it.
type Scene struct {
	Name    string
	World   *physics.World
	Fabrics []*pattern.Fabric
}

// Config controls the window and the physics loop.
type Config struct {
	Width, Height int
	Title         string
	FixedDT       float64 // physics timestep in seconds (e.g. 1/120)
	MaxSubsteps   int     // cap on physics steps per frame (anti spiral-of-death)
	TargetFPS     int32

	// Screenshot, when non-empty, makes the engine render ScreenshotFrames
	// frames, save a PNG to this path and exit. Handy for previews/CI in a
	// headless environment; leave empty for normal interactive use.
	Screenshot       string
	ScreenshotFrames int
}

// DefaultConfig returns a reasonable window/loop configuration.
func DefaultConfig() Config {
	return Config{
		Width:       1280,
		Height:      768,
		Title:       "Crochet Engine",
		FixedDT:     1.0 / 120,
		MaxSubsteps: 8,
		TargetFPS:   60,
	}
}

// Engine owns the window, camera and the active scene. Scenes are produced by
// build functions so they can be rebuilt on demand (reset) and cycled through.
type Engine struct {
	cfg      Config
	builders []func() *Scene
	sceneIdx int
	scene    *Scene

	cam   rl.Camera3D
	orbit orbitCam

	paused    bool
	showLinks bool
	showNodes bool
	windOn    bool
	wind      math3.Vec3

	accumulator float64
}

// New creates an engine over one or more scene builders. Each builder must
// return a fresh world every call, since it is invoked again on reset and
// whenever the scene is selected. Tab cycles through them.
func New(cfg Config, builders ...func() *Scene) *Engine {
	if len(builders) == 0 {
		panic("engine.New: need at least one scene builder")
	}
	e := &Engine{
		cfg:       cfg,
		builders:  builders,
		showLinks: true,
		showNodes: true,
		wind:      math3.V(6, 0, 0),
	}
	e.scene = builders[0]()
	e.orbit = newOrbit(sceneCenter(e.scene))
	return e
}

func (e *Engine) loadScene(i int) {
	n := len(e.builders)
	e.sceneIdx = ((i % n) + n) % n
	e.scene = e.builders[e.sceneIdx]()
	e.orbit.target = sceneCenter(e.scene)
	e.accumulator = 0
}

// Run opens the window and enters the main loop, blocking until the window is
// closed. It must be called from the main goroutine.
func (e *Engine) Run() {
	rl.InitWindow(int32(e.cfg.Width), int32(e.cfg.Height), e.cfg.Title)
	defer rl.CloseWindow()
	if e.cfg.TargetFPS > 0 {
		rl.SetTargetFPS(e.cfg.TargetFPS)
	}

	e.cam = rl.Camera3D{
		Position:   rl.NewVector3(0, 5, 12),
		Target:     v(e.orbit.target),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       45,
		Projection: rl.CameraPerspective,
	}
	e.orbit.apply(&e.cam)

	frame := 0
	for !rl.WindowShouldClose() {
		if e.cfg.Screenshot == "" {
			e.handleInput()
		}
		e.update(float64(rl.GetFrameTime()))
		e.draw()

		if e.cfg.Screenshot != "" {
			frame++
			if frame >= e.cfg.ScreenshotFrames {
				rl.TakeScreenshot(e.cfg.Screenshot)
				return
			}
		}
	}
}

func (e *Engine) handleInput() {
	if rl.IsKeyPressed(rl.KeySpace) || rl.IsKeyPressed(rl.KeyP) {
		e.paused = !e.paused
	}
	if rl.IsKeyPressed(rl.KeyR) {
		e.loadScene(e.sceneIdx)
	}
	if rl.IsKeyPressed(rl.KeyTab) {
		e.loadScene(e.sceneIdx + 1)
	}
	if rl.IsKeyPressed(rl.KeyL) {
		e.showLinks = !e.showLinks
	}
	if rl.IsKeyPressed(rl.KeyN) {
		e.showNodes = !e.showNodes
	}
	e.windOn = rl.IsKeyDown(rl.KeyW)

	e.orbit.handleInput()
	e.orbit.apply(&e.cam)
}

func (e *Engine) update(frame float64) {
	if e.paused {
		e.accumulator = 0
		return
	}
	// Fixed-timestep integration with an accumulator keeps the simulation
	// deterministic and stable regardless of render frame rate.
	e.accumulator += frame
	steps := 0
	for e.accumulator >= e.cfg.FixedDT && steps < e.cfg.MaxSubsteps {
		if e.windOn {
			e.scene.World.AddForceAll(e.wind)
		}
		e.scene.World.Step(e.cfg.FixedDT)
		e.accumulator -= e.cfg.FixedDT
		steps++
	}
	if steps == e.cfg.MaxSubsteps {
		e.accumulator = 0 // fell behind; drop the backlog
	}
}

func (e *Engine) draw() {
	rl.BeginDrawing()
	defer rl.EndDrawing()
	rl.ClearBackground(rl.NewColor(24, 26, 33, 255))

	rl.BeginMode3D(e.cam)
	rl.DrawGrid(24, 1)
	for _, f := range e.scene.Fabrics {
		e.drawFabric(f)
	}
	e.drawColliders()
	rl.EndMode3D()

	e.drawHUD()
}

func (e *Engine) drawFabric(f *pattern.Fabric) {
	w := f.World
	cam := e.orbit.position()
	radius := float32(f.Radius)

	// Thick yarn paths, drawn as a chain of cylinders with rounded joints so
	// rows read as continuous yarn. Each segment is CPU-shaded and may carry a
	// stripe colour.
	for _, s := range f.Strands {
		r := float32(s.Radius)
		for i, seg := range s.Segments() {
			pa := w.Particles[seg[0]].Pos
			pb := w.Particles[seg[1]].Pos
			col := shadeSegment(s.SegColor(i), s.Material, pa, pb, cam)
			rl.DrawCylinderEx(v(pa), v(pb), r, r, 8, col)
		}
		if e.showNodes {
			for i, n := range s.Nodes {
				col := shadePoint(s.SegColor(i), s.Material)
				rl.DrawSphere(v(w.Particles[n].Pos), r, col)
			}
		}
	}

	// Thin structural cross-links between rows.
	if e.showLinks {
		lc := shadePoint(darker(f.Color, 0.6), f.Material)
		lr := radius * 0.4
		for _, l := range f.Links {
			a := v(w.Particles[l[0]].Pos)
			b := v(w.Particles[l[1]].Pos)
			rl.DrawCylinderEx(a, b, lr, lr, 6, lc)
		}
	}

	// Highlight pinned nodes.
	for _, p := range f.Pins {
		rl.DrawSphere(v(w.Particles[p].Pos), radius*1.6, rl.NewColor(240, 220, 90, 255))
	}
}

func (e *Engine) drawColliders() {
	for _, s := range e.scene.World.Spheres {
		rl.DrawSphereEx(v(s.Center), float32(s.Radius), 16, 16, rl.NewColor(90, 100, 120, 120))
	}
}

func (e *Engine) drawHUD() {
	np, nc := 0, 0
	if e.scene.World != nil {
		np = len(e.scene.World.Particles)
		nc = len(e.scene.World.Constraints)
	}
	rl.DrawText(fmt.Sprintf("%s   (%d/%d)", e.scene.Name, e.sceneIdx+1, len(e.builders)),
		12, 10, 22, rl.RayWhite)
	rl.DrawText(fmt.Sprintf("particles: %d   constraints: %d", np, nc),
		12, 38, 18, rl.LightGray)

	state := "running"
	if e.paused {
		state = "PAUSED"
	}
	rl.DrawText(state, 12, 60, 18, rl.NewColor(240, 220, 90, 255))

	help := "drag: orbit   wheel: zoom   right-drag: pan   " +
		"[tab] next  [space] pause  [r] reset  [w] wind  [l] links  [n] nodes"
	rl.DrawText(help, 12, int32(e.cfg.Height)-26, 16, rl.Gray)
	rl.DrawFPS(int32(e.cfg.Width)-90, 10)
}

// --- helpers ---

// v converts a simulation vector to a raylib (float32) vector.
func v(p math3.Vec3) rl.Vector3 {
	return rl.NewVector3(float32(p.X), float32(p.Y), float32(p.Z))
}

// sceneCenter estimates the centroid of all particles so the camera can frame
// the fabric.
func sceneCenter(s *Scene) math3.Vec3 {
	if s == nil || s.World == nil || len(s.World.Particles) == 0 {
		return math3.Zero
	}
	var sum math3.Vec3
	for _, p := range s.World.Particles {
		sum = sum.Add(p.Pos)
	}
	return sum.Scale(1 / float64(len(s.World.Particles)))
}
