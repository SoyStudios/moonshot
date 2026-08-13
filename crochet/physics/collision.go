package physics

import (
	"math"

	"github.com/SoyStudios/moonshot/crochet/math3"
)

// selfCollide keeps the fabric from passing through itself. Every particle is
// treated as a small sphere of radius CollisionRadius; overlapping pairs are
// pushed apart. A uniform spatial hash keeps this near-linear: each particle
// only tests the 27 grid cells around it rather than the whole world.
//
// Particles that are directly joined by a distance constraint are skipped —
// they are *meant* to sit within a stitch-width of each other, and colliding
// them would fight the yarn. Only non-adjacent parts of the fabric repel.
func (w *World) selfCollide() {
	if !w.SelfCollision || w.CollisionRadius <= 0 || len(w.Particles) < 2 {
		return
	}
	minDist := 2 * w.CollisionRadius
	cell := minDist

	// Index lookup so constraints (which hold pointers) can name particles.
	idx := make(map[*Particle]int, len(w.Particles))
	for i, p := range w.Particles {
		idx[p] = i
	}
	// Skip set of directly-linked pairs.
	skip := make(map[[2]int]struct{})
	for _, c := range w.Constraints {
		if dc, ok := c.(*DistanceConstraint); ok {
			a, b := idx[dc.A], idx[dc.B]
			if a > b {
				a, b = b, a
			}
			skip[[2]int{a, b}] = struct{}{}
		}
	}

	key := func(p math3.Vec3) [3]int {
		return [3]int{
			int(math.Floor(p.X / cell)),
			int(math.Floor(p.Y / cell)),
			int(math.Floor(p.Z / cell)),
		}
	}
	buckets := make(map[[3]int][]int, len(w.Particles))
	for i, p := range w.Particles {
		k := key(p.Pos)
		buckets[k] = append(buckets[k], i)
	}

	for i, p := range w.Particles {
		ki := key(p.Pos)
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				for dz := -1; dz <= 1; dz++ {
					nb := buckets[[3]int{ki[0] + dx, ki[1] + dy, ki[2] + dz}]
					for _, j := range nb {
						if j <= i {
							continue
						}
						if _, ok := skip[[2]int{i, j}]; ok {
							continue
						}
						w.resolvePair(p, w.Particles[j], minDist)
					}
				}
			}
		}
	}
}

func (w *World) resolvePair(a, b *Particle, minDist float64) {
	invSum := a.InvMass + b.InvMass
	if invSum == 0 {
		return
	}
	delta := b.Pos.Sub(a.Pos)
	dist := delta.Len()
	if dist >= minDist {
		return
	}
	if dist == 0 {
		// Coincident: nudge deterministically so they can separate.
		b.Pos = b.Pos.Add(math3.V(minDist*0.5, 0, 0))
		return
	}
	// Push each out along the connecting axis by its share of the overlap.
	corr := delta.Scale((dist - minDist) / dist / invSum)
	a.Pos = a.Pos.Add(corr.Scale(a.InvMass))
	b.Pos = b.Pos.Sub(corr.Scale(b.InvMass))
}
