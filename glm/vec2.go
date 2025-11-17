package glm

import "math"

type Vec2[T Numeric] [2]T

func (lhs Vec2[T]) Dot(rhs Vec2[T]) T {
	return (lhs[0] * rhs[0]) + (lhs[1] * rhs[1])
}

func (lhs Vec2[T]) Length() T {
	return T(math.Sqrt(float64(lhs.Dot(lhs))))
}

func (lhs Vec2[T]) Scale(s T) Vec2[T] {
	return Vec2[T]{
		lhs[0] * s,
		lhs[1] * s,
	}
}

func (lhs Vec2[T]) Normalize() Vec2[T] {
	return lhs.Scale(1.0 / lhs.Length())
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

func (lhs Vec2[T]) Mul(rhs Vec2[T]) Vec2[T] {
	return Vec2[T]{
		lhs[0] * rhs[0],
		lhs[1] * rhs[1],
	}
}

func (lhs Vec2[T]) Div(rhs Vec2[T]) Vec2[T] {
	return Vec2[T]{
		lhs[0] / rhs[0],
		lhs[1] / rhs[1],
	}
}

func (lhs Vec2[T]) Min(rhs Vec2[T]) Vec2[T] {
	return Vec2[T]{
		min(lhs[0], rhs[0]),
		min(lhs[1], rhs[1]),
	}
}

func (lhs Vec2[T]) Max(rhs Vec2[T]) Vec2[T] {
	return Vec2[T]{
		max(lhs[0], rhs[0]),
		max(lhs[1], rhs[1]),
	}
}

func (lhs Vec2[T]) Extend(z T) Vec3[T] {
	return Vec3[T]{lhs[0], lhs[1], z}
}

func (lhs Vec2[T]) ToWGPU() [2]float32 {
	return [2]float32{
		float32(lhs[0]),
		float32(lhs[1]),
	}
}

func (lhs Vec2[T]) XY() (x, y T) {
	x = lhs[0]
	y = lhs[1]
	return
}

func (lhs Vec2[T]) ToVec2f() Vec2f {
	return Vec2f{float32(lhs[0]), float32(lhs[1])}
}

func (lhs Vec2[T]) Recip() Vec2[T] {
	return Vec2[T]{
		1.0 / lhs[0],
		1.0 / lhs[1],
	}
}
