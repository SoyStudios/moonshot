# Crochet Engine — Roadmap

Where the engine is headed. Roughly ordered by how much each unlocks, not by
difficulty. Nothing here is committed to a date; it's a menu of the next
worthwhile steps.

## Modelling fidelity

- **Stitch rendering.** ✅ Fake per-stitch "V" geometry (see
  `engine/stitchrender.go`) makes fabric read as crochet without extra physics.
  Follow-ups: distinct silhouettes per stitch type (a `dc` V taller and with a
  visible post, a bobble as a cluster), and the front-loop/back-loop "bar" that
  sits between rows.
- **Stitch posts.** Subdivide tall stitches (`dc`, `tr`) into flexible legs so
  the `Stitch` gauge drives real *physics* geometry — a post that can bend and
  lean — not just row spacing and a drawn V. Ties into self-collision (posts are
  collidable) and makes textured stitch patterns (front/back post, cables)
  possible.
- **Increase / decrease placement.** ✅ Partly — `pattern.Build` anchors every
  stitch to the specific stitch(es) it's worked into, so inc/dec topology is
  correct and the round-by-round counts shape the silhouette. Still to do:
  *positional* placement (increases clustered on one side) for asymmetric,
  sculpted pieces — today each round is laid out as an even ring.
- **Stitch texture library.** Bobbles, popcorns, front/back loop only, clusters,
  shells — each as a small builder that emits nodes + bonds, composable into
  rows and rounds.
- **Seams and assembly.** Join finished pieces (sew a limb to a body) by linking
  their edge stitches, so multi-part amigurumi hold together under physics.

## Authoring

- **Written-pattern parser.** ✅ Done — `pattern.Parse` reads standard notation
  (`R1: 6 sc in magic ring`, `R2: inc x6`, `R3: (sc, inc) x6`, ranges, fill
  rounds, `sc2tog`) into a `Pattern`, and `pattern.Build` turns it into a
  `Fabric`. Follow-ups: foundation-chain (flat) starts, colour changes and
  loop-only (BLO/FLO) stitches in the notation, and round-level validation
  (warn when a round's ops don't consume the round below).
- **Pattern DSL / builder API.** A fluent Go API (`p.Round().Sc(6).Inc(6)`) over
  the same `Op`/`Round` model the parser produces, for authoring in code.
- **Gauge & yarn-weight presets.** Map real yarn weights (DK, worsted, aran) and
  hook sizes to gauge/radius so pieces come out at believable proportions.

## Simulation

- **Better self-collision.** Segment/segment (capsule) collision instead of
  point spheres, and continuous collision for fast motion, so thin yarn layers
  never tunnel through each other.
- **Stuffing as true volume.** Replace the centroid-pressure approximation with
  a volume-preserving soft-body model for non-convex shapes (limbs, snouts).
- **Anisotropic yarn.** Twist and ply so yarn resists bending differently along
  vs. across the strand; enables realistic curling of stockinette-like edges.
- **Wind / interaction forces.** Grab-and-drag with the mouse, pinning at
  runtime, cutting yarn (tearable constraints already exist).

## Rendering

- **GPU shader lighting.** ✅ Done — a matte fabric shader (wrapped diffuse, dim
  broad sheen, helical **ply** relief and micro-roughness) plus a **shell-fur
  fibre halo** for a fuzzy silhouette. The fuzz noise is **baked into a tiling
  texture** and the instance matrices are built directly from the direction
  vector (no per-frame `acos`/rotation compose), which roughly doubled the
  software-render framerate (see `engine/yarnshader.go`). Follow-ups: multiple
  lights, shadow mapping, and frustum culling of off-screen stitches.
- **Batching & LOD.** ✅ Instanced draws grouped by colour, plus per-stitch
  level of detail (full V → straight V → single bar by apparent size). Follow-up:
  a baked per-fabric mesh with vertex colours for a true single draw call, and
  frustum culling of off-screen stitches.
- **Real tube meshes.** Generate swept tube geometry with proper normals and UVs
  for texture (twist lines, ply) instead of stacked cylinders.
- **Export.** Save a posed piece to a mesh (OBJ/glTF) for rendering elsewhere or
  3-D printing a stitch guide.

## Tooling & UX

- **Scene selector / inspector UI.** Pick patterns, tweak yarn and physics
  parameters live, scrub the simulation.
- **Deterministic replay & headless render.** Extend the existing screenshot
  hook into short turntable clips for docs and regression snapshots.
- **Benchmarks.** Track solver cost as stitch counts grow; keep the spatial hash
  honest.
