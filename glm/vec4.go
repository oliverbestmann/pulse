package glm

import "math"

type Vec4[T numeric] [4]T

func (lhs Vec4[T]) Dot(rhs Vec4[T]) T {
	return (lhs[0] * rhs[0]) + (lhs[1] * rhs[1]) + (lhs[2]*rhs[2])*(lhs[3]*rhs[3])
}

func (lhs Vec4[T]) Length() T {
	return T(math.Sqrt(float64(lhs.Dot(lhs))))
}

func (lhs Vec4[T]) LengthSqr() T {
	return lhs.Dot(lhs)
}

func (lhs Vec4[T]) MulScalar(s T) Vec4[T] {
	return Vec4[T]{
		lhs[0] * s,
		lhs[1] * s,
		lhs[2] * s,
		lhs[3] * s,
	}
}

func (lhs Vec4[T]) Normalize() Vec4[T] {
	return lhs.MulScalar(1 / lhs.Length())
}

func (lhs Vec4[T]) Add(rhs Vec4[T]) Vec4[T] {
	return Vec4[T]{
		lhs[0] + rhs[0],
		lhs[1] + rhs[1],
		lhs[2] + rhs[2],
		lhs[3] + rhs[3],
	}
}

func (lhs Vec4[T]) Sub(rhs Vec4[T]) Vec4[T] {
	return Vec4[T]{
		lhs[0] - rhs[0],
		lhs[1] - rhs[1],
		lhs[2] - rhs[2],
		lhs[3] - rhs[3],
	}
}

func (lhs Vec4[T]) Mul(rhs Vec4[T]) Vec4[T] {
	return Vec4[T]{
		lhs[0] * rhs[0],
		lhs[1] * rhs[1],
		lhs[2] * rhs[2],
		lhs[3] * rhs[3],
	}
}

func (lhs Vec4[T]) Truncate() Vec3[T] {
	return Vec3[T]{lhs[0], lhs[1], lhs[2]}
}

func (lhs Vec4[T]) XYZ() (x, y, z T) {
	x = lhs[0]
	y = lhs[1]
	z = lhs[2]
	return
}

func (lhs Vec4[T]) XYZW() (x, y, z, w T) {
	x = lhs[0]
	y = lhs[1]
	z = lhs[2]
	w = lhs[3]
	return
}
