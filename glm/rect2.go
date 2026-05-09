package glm

import (
	"fmt"
)

type Rect2f = Rect2[float32]
type Rect2u = Rect2[uint32]
type Rect2uh = Rect2[uint16]

type Rect2[T Numeric] struct {
	Min Vec2[T]
	Max Vec2[T]
}

func RectFromSize[T Numeric](pos Vec2[T], size Vec2[T]) Rect2[T] {
	return RectFromPoints[T](pos, pos.Add(size))
}

func RectFromXYWH[T Numeric](x, y, w, h T) Rect2[T] {
	pos := Vec2[T]{x, y}
	size := Vec2[T]{w, h}
	return RectFromSize[T](pos, size)
}

func RectFromPoints[T Numeric](a, b Vec2[T]) Rect2[T] {
	return Rect2[T]{
		Min: Vec2[T]{
			min(a[0], b[0]),
			min(a[1], b[1]),
		},
		Max: Vec2[T]{
			max(a[0], b[0]),
			max(a[1], b[1]),
		},
	}
}

func (r Rect2[T]) Extend(point Vec2[T]) Rect2[T] {
	minX := min(r.Min[0], point[0])
	minY := min(r.Min[1], point[1])

	maxX := max(r.Max[0], point[0])
	maxY := max(r.Max[1], point[1])

	return Rect2[T]{
		Min: Vec2[T]{minX, minY},
		Max: Vec2[T]{maxX, maxY},
	}
}

func (r Rect2[T]) Union(other Rect2[T]) Rect2[T] {
	return r.Extend(other.Min).Extend(other.Max)
}

func (r Rect2[T]) Contains(other Rect2[T]) bool {
	return r.Min[0] <= other.Min[0] && r.Min[1] <= other.Min[1] &&
		r.Max[0] >= other.Max[0] && r.Max[1] >= other.Max[1]
}

func (r Rect2[T]) Center() Vec2[T] {
	return r.Min.Add(r.Max).Div(Vec2[T]{2, 2})
}

func (r Rect2[T]) Offset() Vec2[T] {
	return r.Min
}

func (r Rect2[T]) Size() Vec2[T] {
	return r.Max.Sub(r.Min)
}

func (r Rect2[T]) Width() T {
	return r.Max[0] - r.Min[0]
}

func (r Rect2[T]) Height() T {
	return r.Max[1] - r.Min[1]
}

func (r Rect2[T]) XYWH() (T, T, T, T) {
	x, y := r.Min.XY()
	w, h := r.Size().XY()
	return x, y, w, h
}

func (r Rect2[T]) String() string {
	x, y, w, h := r.XYWH()
	return fmt.Sprintf("Rect(x=%v, y=%v, w=%v, h=%v)", x, y, w, h)
}
