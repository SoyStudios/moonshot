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
| `M` | toggle stitch view / wire view |
| `F` | toggle fuzzy fibre halo |
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

By default the renderer draws **faked stitch geometry**: each stitch becomes a
small bowed "V" — the shape a knit/crochet stitch's top loops make — sized from
the stitch gauge and oriented by the *live* fabric surface. On a flat piece the
neighbours give the frame directly (row × column → normal); on a piece worked in
the round the neighbour cross product jitters, so the normal is taken from the
smooth outward direction instead (point − centroid for a ball, the same with the
vertical axis removed for a tube) and the row/column axes are rebuilt orthogonal
to it — which keeps stitches on curved surfaces tidy. Rows of these Vs interlock
into recognisable fabric that deforms with the simulation. It's purely cosmetic
— the physics still sees one particle per stitch. Press `M` to switch to the
**wire view**, which shows the raw yarn paths and structural links.

Lighting runs on the **GPU** through a GLSL material tuned to look like yarn
rather than plastic. Diffuse is soft and wrapped (fibres scatter light around
the terminator, so there's no hard shadow edge); the highlight is a dim, broad
fabric sheen plus a faint grazing rim, never a tight hot spot — the material
**sheen** uniform scales it, so matte wool and glossy silk still differ. On top
of that the surface is roughened two ways: a helical **ply** pattern (from each
fragment's angle around the tube and distance along it) carves twisted-strand
ridges and grooves, and a **baked noise texture** (generated once and sampled by
UV, with a per-instance offset so it doesn't visibly tile) jitters the normal
for a fuzzy, light-scattering micro-surface. Around all of it sits a **fibre
halo**: the cylinders are re-drawn as a couple of translucent shells pushed out
along the normal, with the same noise masking fewer fragments the further out it
goes, so sparse fibre tips fuzz the silhouette like real wool (toggle with `F`,
or start with `CROCHET_FUZZ=0` for a cheaper draw). Stitches
carry a **stripe** palette for self-striping colourwork. (If the shader can't
compile on a limited driver, the renderer falls back to flat cylinders so the
demo still runs.)

Drawing is **instanced and batched**: every cylinder/sphere is accumulated as a
per-instance transform, grouped by colour+material, and each group is issued as
a single `DrawMeshInstanced` call — so a whole fabric costs a handful of draw
calls instead of thousands. Each stitch also picks a **level of detail** from
how large it appears on screen: a full bowed V up close, a straight V at middle
distance, a single bar when far away, cutting geometry where it wouldn't be
seen.

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

## Writing patterns

The `pattern` package also reads crochet the way it's actually written.
`pattern.Parse` turns standard notation into a `Pattern`, and `pattern.Build`
materialises it — anchoring every stitch to the specific stitch(es) it's worked
into, so increases (two into one) and decreases (one across two) have the right
topology rather than a nearest-neighbour guess:

```go
p := pattern.MustParse(`
    R1: 6 sc in magic ring
    R2: inc x6            (12)
    R3: (sc, inc) x6      (18)
    R4: (2 sc, inc) x6    (24)
    R5-9: sc              (24)
    R10: (2 sc, dec) x6   (18)
    R11: (sc, dec) x6     (12)
    R12: dec x6           (6)
`)
fabric := pattern.Build(world, p, pattern.BuildConfig{
    Gauge: 0.42, Center: math3.V(0, 3, 0), CloseTop: true, Yarn: yarn.DefaultConfig(),
})
fabric.Stuff(0.004) // stuff it
```

The parser handles round labels and ranges (`R5-9:`), repeats (`(sc, inc) x6`,
`inc x6`, `x`/`*`/`×`), leading counts (`2 sc`), fill rounds (a bare `sc` means
"one in each stitch"), the `sc2tog`/`dec` decrease spellings, and ignores
stitch-count annotations like `(18)` and filler like "in each st". Each round's
radius follows its stitch count, so the increases and decreases sculpt the
silhouette; the demo's amigurumi ball is built straight from the text above.

## Extending

- New shapes: add a builder to `pattern` returning a `*Fabric`, drive `Revolve`
  with a new count sequence, or just write a pattern for `Parse`/`Build`.
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
