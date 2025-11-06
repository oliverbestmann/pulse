package glm

import "math"

type Vec3[T numeric] [3]T

func (lhs Vec3[T]) Dot(rhs Vec3[T]) T {
	return (lhs[0] * rhs[0]) + (lhs[1] * rhs[1]) + (lhs[2] * rhs[2])
}

func (lhs Vec3[T]) Magnitude() T {
	return T(math.Sqrt(float64(lhs.Dot(lhs))))
}

func (lhs Vec3[T]) MulScalar(s T) Vec3[T] {
	return Vec3[T]{
		lhs[0] * s,
		lhs[1] * s,
		lhs[2] * s,
	}
}

func (lhs Vec3[T]) Normalize() Vec3[T] {
	return lhs.MulScalar(1 / lhs.Magnitude())
}

func (lhs Vec3[T]) Cross(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[1]*rhs[2] - rhs[1]*lhs[2],
		lhs[2]*rhs[0] - rhs[2]*lhs[0],
		lhs[0]*rhs[1] - rhs[0]*lhs[1],
	}
}

func (lhs Vec3[T]) Add(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[0] + rhs[0],
		lhs[1] + rhs[1],
		lhs[2] + rhs[2],
	}
}

func (lhs Vec3[T]) Sub(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[0] - rhs[0],
		lhs[1] - rhs[1],
		lhs[2] - rhs[2],
	}
}

func (lhs Vec3[T]) Mul(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[0] * rhs[0],
		lhs[1] * rhs[1],
		lhs[2] * rhs[2],
	}
}

func (lhs Vec3[T]) Div(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[0] / rhs[0],
		lhs[1] / rhs[1],
		lhs[2] / rhs[2],
	}
}

func (lhs Vec3[T]) Truncate() Vec2[T] {
	return Vec2[T]{lhs[0], lhs[1]}
}

func (lhs Vec3[T]) Extend(w T) Vec4[T] {
	return Vec4[T]{lhs[0], lhs[1], lhs[2], w}
}

func (lhs Vec3[T]) XY() (x, y T) {
	x = lhs[0]
	y = lhs[1]
	return
}

func (lhs Vec3[T]) XYZ() (x, y, z T) {
	x = lhs[0]
	y = lhs[1]
	z = lhs[2]
	return
}
