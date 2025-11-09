package pulse

import (
	"github.com/oliverbestmann/go3d/glm"
	"golang.org/x/exp/constraints"
)

type numeric interface {
	constraints.Integer | constraints.Float
}

type Rectangle2f = Rectangle2[float32]
type Rectangle2u = Rectangle2[uint32]

type Rectangle2[T numeric] struct {
	Min glm.Vec2[T]
	Max glm.Vec2[T]
}

func RectangleFromSize[T numeric](pos glm.Vec2[T], size glm.Vec2[T]) Rectangle2[T] {
	return RectangleFromPoints[T](pos, pos.Add(size))
}

func RectangleFromPoints[T numeric](a, b glm.Vec2[T]) Rectangle2[T] {
	return Rectangle2[T]{
		Min: glm.Vec2[T]{
			min(a[0], b[0]),
			min(a[1], b[1]),
		},
		Max: glm.Vec2[T]{
			max(a[0], b[0]),
			max(a[1], b[1]),
		},
	}
}

func (r Rectangle2[T]) Extend(point glm.Vec2[T]) Rectangle2[T] {
	minX := min(r.Min[0], point[0])
	minY := min(r.Min[1], point[1])

	maxX := max(r.Max[0], point[0])
	maxY := max(r.Max[1], point[1])

	return Rectangle2[T]{
		Min: glm.Vec2[T]{minX, minY},
		Max: glm.Vec2[T]{maxX, maxY},
	}
}

func (r Rectangle2[T]) Union(other Rectangle2[T]) Rectangle2[T] {
	return r.Extend(other.Min).Extend(other.Max)
}

func (r Rectangle2[T]) Center() glm.Vec2[T] {
	return r.Min.Add(r.Max).Div(glm.Vec2[T]{2, 2})
}

func (r Rectangle2[T]) Size() glm.Vec2[T] {
	return r.Max.Sub(r.Min)
}

func (r Rectangle2[T]) Width() T {
	return r.Max[0] - r.Min[0]
}

func (r Rectangle2[T]) Height() T {
	return r.Max[1] - r.Min[1]
}

func (r Rectangle2[T]) XYWH() (T, T, T, T) {
	x, y := r.Min.XY()
	w, h := r.Size().XY()
	return x, y, w, h
}
