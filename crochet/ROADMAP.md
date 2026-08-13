# Crochet Engine — Roadmap

Where the engine is headed. Roughly ordered by how much each unlocks, not by
difficulty. Nothing here is committed to a date; it's a menu of the next
worthwhile steps.

## Modelling fidelity

- **Stitch posts.** Subdivide tall stitches (`dc`, `tr`) into flexible legs so
  the `Stitch` gauge drives real geometry — a post that can bend and lean — not
  just row spacing. Ties into self-collision (posts are collidable) and makes
  textured stitch patterns (front/back post, cables) possible.
- **Increase / decrease placement.** Move beyond surfaces of revolution: place
  increases and decreases at chosen stitches so amigurumi can be sculpted
  (heads, snouts, limbs, tapered shapes) rather than only bell-curve counts.
- **Stitch texture library.** Bobbles, popcorns, front/back loop only, clusters,
  shells — each as a small builder that emits nodes + bonds, composable into
  rows and rounds.
- **Seams and assembly.** Join finished pieces (sew a limb to a body) by linking
  their edge stitches, so multi-part amigurumi hold together under physics.

## Authoring

- **Written-pattern parser.** Read standard written patterns
  (`R1: 6 sc in magic ring`, `R2: inc ×6 (12)`, `R3: (sc, inc) ×6 (18)`) straight
  into a `Fabric`. This is the bridge from "hand-coded scenes" to "paste a real
  pattern and watch it."
- **Pattern DSL / builder API.** A fluent Go API (`p.Round().Sc(6)`,
  `p.Round().Inc(6)`) that the text parser targets, usable directly in code too.
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

- **GPU shader lighting.** ✅ Done — per-fragment diffuse + specular in a GLSL
  material (see `engine/yarnshader.go`). Follow-ups: multiple lights, shadow
  mapping, ambient occlusion in the stitch valleys, and a fuzzy/fibre "halo" for
  a yarn-like silhouette.
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
