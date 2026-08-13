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
| Draped swatch | a sheet dropping over a spherical form and folding, with self-collision |
| Stuffed ball | an amigurumi ball sealed at both poles, inflated by internal stuffing |
| Striped beanie | a hat worked in the round with self-striping colourwork |
| Materials | two identical swatches — matte wool vs glossy silk — under the same light |
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
- **Self-collision** treats each particle as a small sphere and pushes
  overlapping, non-adjacent parts apart, so folded fabric doesn't pass through
  itself. A uniform spatial hash keeps it near-linear, and directly-linked
  stitches are skipped so collision never fights the yarn.
- **Stuffing** is a `PressureConstraint`: it pushes a closed shape's particles
  outward from their centroid each solve, and the yarn tension balances it, so a
  sealed amigurumi inflates to a plump equilibrium.

The renderer draws each yarn segment as a cylinder with rounded joints, and the
inter-row bonds as thin posts, so rows read as continuous crochet. Lighting runs
on the **GPU**: unit cylinder/sphere meshes are drawn through a GLSL material
that does per-fragment diffuse + specular shading, with the material **sheen**
as a uniform, so matte wool and glossy silk look different. Strands can carry a
**stripe** palette for self-striping colourwork. (If the shader can't compile on
a limited driver, the renderer falls back to flat cylinders so the demo still
runs.)

## How a crochet piece is built

A `pattern` builder lays out one particle per stitch, threads a continuous yarn
path through them (boustrophedon for flat rows, closed loops for rounds), and
adds cross-links between each stitch and the stitch it was "worked into" in the
previous row. That single idea — a stitch grid bonded by yarn — produces
swatches, tubes and increasing discs.

`Revolve` generalises it: hand it a sequence of per-round **stitch counts** (the
way amigurumi is actually written — `6, 12, 18, … , 12, 6`) and it lays out a
surface of revolution. Each round's circumference follows its count, so
increases bulge the shape out and decreases pull it in; the radius and rise are
derived so the slant between rounds matches the stitch height. `SphereCounts`
generates the bell-curve counts for a ball; feed a ramp for a cone. `CloseTop`
/`CloseBottom` cinch the poles so a `Stuff` pressure constraint can inflate a
sealed shape.

Each `pattern.Stitch` carries a **gauge** (`Def{Width, Height}`) for
`sc`/`hdc`/`dc`/`tr`, so taller stitches make taller rows.

## Extending

- New shapes: add a builder to `pattern` returning a `*Fabric`, or drive
  `Revolve` with a new count sequence.
- New stitches: extend `pattern.Stitch` / its `Def` gauge.
- New physics: implement `physics.Constraint` (like the pressure/stuffing
  constraint) and add it to the world.
- New looks: set a `yarn.Material` (sheen/ambient) or a `Stripe` palette on the
  yarn config.

## Status

Functional and growing: the yarn solver (with self-collision and stuffing
pressure), stitch-gauge-aware builders (swatch, tube, disc, and the
count-driven `Revolve`), CPU-shaded material/colourwork rendering, a 3-D viewer
and a headless test suite all work. Next steps worth exploring are proper stitch
posts (subdividing tall stitches into flexible legs), reading written patterns
from text, and increase/decrease placement for sculpted amigurumi.
