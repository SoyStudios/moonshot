package math3

import (
	"math"
	"testing"
)

const eps = 1e-9

func approx(a, b float64) bool { return math.Abs(a-b) < eps }

func TestAddSubScale(t *testing.T) {
	a := V(1, 2, 3)
	b := V(4, 5, 6)
	if got := a.Add(b); got != (Vec3{5, 7, 9}) {
		t.Fatalf("Add = %+v", got)
	}
	if got := b.Sub(a); got != (Vec3{3, 3, 3}) {
		t.Fatalf("Sub = %+v", got)
	}
	if got := a.Scale(2); got != (Vec3{2, 4, 6}) {
		t.Fatalf("Scale = %+v", got)
	}
}

func TestDotCross(t *testing.T) {
	x := V(1, 0, 0)
	y := V(0, 1, 0)
	if !approx(x.Dot(y), 0) {
		t.Fatalf("perpendicular dot should be 0")
	}
	if got := x.Cross(y); got != (Vec3{0, 0, 1}) {
		t.Fatalf("x cross y should be z, got %+v", got)
	}
}

func TestLenNormalize(t *testing.T) {
	v := V(3, 4, 0)
	if !approx(v.Len(), 5) {
		t.Fatalf("Len = %v, want 5", v.Len())
	}
	n := v.Normalize()
	if !approx(n.Len(), 1) {
		t.Fatalf("normalized length = %v", n.Len())
	}
}

func TestNormalizeZeroSafe(t *testing.T) {
	if got := Zero.Normalize(); got != Zero {
		t.Fatalf("normalizing zero should stay zero, got %+v", got)
	}
}

func TestLerp(t *testing.T) {
	a := V(0, 0, 0)
	b := V(10, 20, 30)
	mid := Lerp(a, b, 0.5)
	if mid != (Vec3{5, 10, 15}) {
		t.Fatalf("Lerp mid = %+v", mid)
	}
}
