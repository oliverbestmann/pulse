package glm

import "math"

type Vec2[T numeric] [2]T

func (lhs Vec2[T]) Dot(rhs Vec2[T]) T {
	return (lhs[0] * rhs[0]) + (lhs[1] * rhs[1])
}

func (lhs Vec2[T]) Magnitude() T {
	return T(math.Sqrt(float64(lhs.Dot(lhs))))
}

func (lhs Vec2[T]) MulScalar(s T) Vec2[T] {
	return Vec2[T]{
		lhs[0] * s,
		lhs[1] * s,
	}
}

func (lhs Vec2[T]) Normalize() Vec2[T] {
	return lhs.MulScalar(1 / lhs.Magnitude())
}

func (lhs Vec2[T]) Add(rhs Vec2[T]) Vec2[T] {
	return Vec2[T]{
		lhs[0] + rhs[0],
		lhs[1] + rhs[1],
	}
}

func (lhs Vec2[T]) Sub(rhs Vec2[T]) Vec2[T] {
	return Vec2[T]{
		lhs[0] - rhs[0],
		lhs[1] - rhs[1],
	}
}

func (lhs Vec2[T]) Extend(z T) Vec3[T] {
	return Vec3[T]{lhs[0], lhs[1], z}
}

func (lhs Vec2[T]) XY() (x, y T) {
	x = lhs[0]
	y = lhs[1]
	return
}
