package glm

import (
	"unsafe"
)

type Mat3[T numeric] [9]T

func IdentityMat3[T numeric]() Mat3[T] {
	return Mat3[T]{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

func TranslationMat3[T numeric](x, y T) Mat3[T] {
	return Mat3[T]{
		1, 0, 0,
		0, 1, 0,
		x, y, 1,
	}
}

func RotationMat3[T numeric](angle Rad) Mat3[T] {
	s, c := fastSincos(angle)

	return Mat3[T]{
		T(c), T(s), 0,
		-T(s), T(c), 0,
		0, 0, 1,
	}
}

func ScaleMat3[T numeric](x, y T) Mat3[T] {
	return Mat3[T]{
		x, 0, 0,
		0, y, 0,
		0, 0, 1,
	}
}

func (lhs Mat3[T]) Rotate(angle Rad) Mat3[T] {
	s, c := fastSincos(angle)

	rhs := Mat3[T]{
		T(c), T(s), 0,
		-T(s), T(c), 0,
		0, 0, 1,
	}

	return lhs.Mul(rhs)
}

func (lhs Mat3[T]) Scale(x, y T) Mat3[T] {
	rhs := Mat3[T]{
		x, 0, 0,
		0, y, 0,
		0, 0, 1,
	}

	return lhs.Mul(rhs)
}

func (lhs Mat3[T]) Translate(x, y T) Mat3[T] {
	rhs := Mat3[T]{
		1, 0, 0,
		0, 1, 0,
		x, y, 1,
	}

	return lhs.Mul(rhs)
}

func (lhs Mat3[T]) Mul(rhs Mat3[T]) Mat3[T] {
	return Mat3[T]{
		lhs[0]*rhs[0] + lhs[3]*rhs[1] + lhs[6]*rhs[2],
		lhs[1]*rhs[0] + lhs[4]*rhs[1] + lhs[7]*rhs[2],
		lhs[2]*rhs[0] + lhs[5]*rhs[1] + lhs[8]*rhs[2],
		lhs[0]*rhs[3] + lhs[3]*rhs[4] + lhs[6]*rhs[5],
		lhs[1]*rhs[3] + lhs[4]*rhs[4] + lhs[7]*rhs[5],
		lhs[2]*rhs[3] + lhs[5]*rhs[4] + lhs[8]*rhs[5],
		lhs[0]*rhs[6] + lhs[3]*rhs[7] + lhs[6]*rhs[8],
		lhs[1]*rhs[6] + lhs[4]*rhs[7] + lhs[7]*rhs[8],
		lhs[2]*rhs[6] + lhs[5]*rhs[7] + lhs[8]*rhs[8],
	}
}

func (lhs Mat3[T]) Transform(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[0]*rhs[0] + lhs[3]*rhs[1] + lhs[6]*rhs[2],
		lhs[1]*rhs[0] + lhs[4]*rhs[1] + lhs[7]*rhs[2],
		lhs[2]*rhs[0] + lhs[5]*rhs[1] + lhs[8]*rhs[2],
	}
}

func (lhs Mat3[T]) IsZero() bool {
	return lhs == Mat3[T]{}
}

func (lhs Mat3[T]) Transpose() Mat3[T] {
	// original
	// 0  1  2
	// 3  4  5
	// 6  7  8

	// transposed
	// 0  3  6
	// 1  4  7
	// 2  5  8

	return Mat3[T]{
		lhs[0], lhs[3], lhs[6],
		lhs[1], lhs[4], lhs[7],
		lhs[2], lhs[5], lhs[8],
	}
}

func (lhs Mat3[T]) Row(i int) Vec3[T] {
	return Vec3[T]{
		lhs[i+0],
		lhs[i+3],
		lhs[i+6],
	}
}

func (lhs Mat3[T]) Columns() [3]Vec3[T] {
	return *(*[3]Vec3[T])(unsafe.Pointer(&lhs))
}
