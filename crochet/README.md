# Crochet Engine

A small 3-D engine, written in Go on top of
[raylib-go](https://github.com/gen2brain/raylib-go), for **simulating and
visualizing crochet works as physical yarn**.

Crochet is just one continuous strand of yarn looped through itself. This engine
models that literally: yarn is a chain of point masses held together by
distance constraints, stitches bond neighbouring loops, and the whole assembly
relaxes under gravity into the shape a real piece would take.

## Run the demo

```sh
go run ./crochet/cmd/crochet
```

A window opens with several crochet pieces you can cycle through:

| Scene | What it shows |
|-------|---------------|
| Hanging swatch | a flat sampler pinned along its top edge, hanging and swaying |
| Draped swatch | a sheet of crochet dropping over a spherical form and folding |
| Beanie tube | crochet worked in the round into a cylinder, pinned at the crown |
| Amigurumi disc | a magic-ring circle with increases that ruffles into a bowl |

### Controls

| Input | Action |
|-------|--------|
| left-drag | orbit camera |
| wheel | zoom |
| right-drag | pan |
| `Tab` | next scene |
| `Space` / `P` | pause / resume |
| `R` | reset current scene |
| `W` (hold) | wind gust |
| `L` | toggle structural cross-links |
| `N` | toggle stitch nodes |

## Architecture

The engine is layered so that everything except rendering is pure Go and can be
unit-tested without a display or cgo:

```
cmd/crochet   demo: builds scenes, opens the window
   │
engine        raylib window, orbit camera, tube rendering, fixed-timestep loop   (cgo)
   │
pattern       crochet builders: Swatch, Tube, Disc → particles + constraints
   │
yarn          Strand primitive: threads particles and wires their constraints
   │
physics       Verlet integrator + position-based distance/bending constraints,
   │           ground & sphere colliders
math3         dependency-free Vec3 math
```

Only `engine` (and the demo) import raylib. `math3`, `physics`, `yarn` and
`pattern` have no C dependencies, so `go test ./crochet/math3/...
./crochet/physics/... ./crochet/pattern/...` runs headlessly.

## The physics

Yarn is inextensible but floppy — a perfect fit for **Verlet integration with
Position Based Dynamics**:

- **Particles** store a position and a *previous* position; velocity is the
  difference between them. Pinned particles have infinite mass and never move.
- **Distance constraints** pull two particles back to a rest separation each
  step. Chained along a line they make yarn; cross-linked they make fabric. The
  same constraint, applied between every *other* particle at a lower stiffness,
  provides **bending stiffness** so the yarn resists sharp kinks. Constraints
  can also **tear** past a stretch threshold.
- The **world** integrates once, then relaxes all constraints and resolves
  collisions over several iterations — more iterations means stiffer, less
  stretchy yarn. A ground plane (with friction) and sphere colliders let fabric
  drape over forms.

The renderer draws each yarn segment as a tapered cylinder with rounded joints,
and the inter-row bonds as thin posts, so rows read as continuous crochet.

## How a crochet piece is built

A `pattern` builder lays out one particle per stitch, threads a continuous yarn
path through them (boustrophedon for flat rows, closed loops for rounds), and
adds cross-links between each stitch and the stitch it was "worked into" in the
previous row. That single idea — a stitch grid bonded by yarn — produces
swatches, tubes and increasing discs, and is the foundation for richer stitch
types and patterns later.

## Extending

- New shapes: add a builder to `pattern` returning a `*Fabric`.
- New stitches: `pattern.Stitch` already carries relative heights (`sc`, `hdc`,
  `dc`, `tr`, …); use them to vary row spacing.
- New physics: implement `physics.Constraint` (e.g. shear or volume
  constraints) and add it to the world.

## Status

Early but functional: the yarn solver, three crochet builders, a 3-D viewer and
a headless test suite all work. Next steps worth exploring are self-collision
between stitches (so fabric doesn't pass through itself), stuffing pressure for
amigurumi, and reading real written patterns.
