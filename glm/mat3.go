package glm

// Mat3 is a 3x3 matrix.
// The default value is the identity matrix.
type Mat3[T numeric] struct {
	values [3][3]T
}

func Mat3Of[T numeric](m [3][3]T) Mat3[T] {
	const one = 1
	const zer = 0

	return Mat3[T]{
		// offset values
		values: [3][3]T{
			{m[0][0] - one, m[0][1] - zer, m[0][2] - zer},
			{m[1][0] - zer, m[1][1] - one, m[1][2] - zer},
			{m[2][0] - zer, m[2][1] - zer, m[2][2] - one},
		},
	}
}

func IdentityMat3[T numeric]() Mat3[T] {
	return Mat3[T]{}
}

func TranslationMat3[T numeric](x, y T) Mat3[T] {
	return Mat3Of([3][3]T{
		{1, 0, 0},
		{0, 1, 0},
		{x, y, 1},
	})
}

func RotationMat3[T numeric](angle Rad) Mat3[T] {
	s, c := fastSincos(angle)

	return Mat3Of([3][3]T{
		{T(c), T(s), 0},
		{-T(s), T(c), 0},
		{0, 0, 1},
	})
}

func ScaleMat3[T numeric](x, y T) Mat3[T] {
	return Mat3Of([3][3]T{
		{x, 0, 0},
		{0, y, 0},
		{0, 0, 1},
	})
}

func (lhs Mat3[T]) Rotate(angle Rad) Mat3[T] {
	rhs := RotationMat3[T](angle)
	return lhs.Mul(rhs)
}

func (lhs Mat3[T]) Scale(x, y T) Mat3[T] {
	rhs := ScaleMat3(x, y)
	return lhs.Mul(rhs)
}

func (lhs Mat3[T]) Translate(x, y T) Mat3[T] {
	rhs := TranslationMat3(x, y)
	return lhs.Mul(rhs)
}

func (lhs Mat3[T]) Mul(rhs Mat3[T]) Mat3[T] {
	return Mat3Of([3][3]T{
		{
			lhs.m00()*rhs.m00() + lhs.m10()*rhs.m01() + lhs.m20()*rhs.m02(),
			lhs.m01()*rhs.m00() + lhs.m11()*rhs.m01() + lhs.m21()*rhs.m02(),
			lhs.m02()*rhs.m00() + lhs.m12()*rhs.m01() + lhs.m22()*rhs.m02(),
		},
		{
			lhs.m00()*rhs.m10() + lhs.m10()*rhs.m11() + lhs.m20()*rhs.m12(),
			lhs.m01()*rhs.m10() + lhs.m11()*rhs.m11() + lhs.m21()*rhs.m12(),
			lhs.m02()*rhs.m10() + lhs.m12()*rhs.m11() + lhs.m22()*rhs.m12(),
		},
		{
			lhs.m00()*rhs.m20() + lhs.m10()*rhs.m21() + lhs.m20()*rhs.m22(),
			lhs.m01()*rhs.m20() + lhs.m11()*rhs.m21() + lhs.m21()*rhs.m22(),
			lhs.m02()*rhs.m20() + lhs.m12()*rhs.m21() + lhs.m22()*rhs.m22(),
		},
	})
}

func (lhs Mat3[T]) Transform(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		(lhs.m00())*rhs[0] + (lhs.m10())*rhs[1] + (lhs.m20())*rhs[2],
		(lhs.m01())*rhs[0] + (lhs.m11())*rhs[1] + (lhs.m21())*rhs[2],
		(lhs.m02())*rhs[0] + (lhs.m12())*rhs[1] + (lhs.m22())*rhs[2],
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

	return Mat3Of([3][3]T{
		{lhs.m00(), lhs.m10(), lhs.m20()},
		{lhs.m01(), lhs.m11(), lhs.m21()},
		{lhs.m02(), lhs.m12(), lhs.m22()},
	})
}

func (lhs Mat3[T]) Row(i int) Vec3[T] {
	if i == 0 {
		return Vec3[T]{
			lhs.m00(),
			lhs.m10(),
			lhs.m20(),
		}
	}
	if i == 1 {
		return Vec3[T]{lhs.m01(), lhs.m11(), lhs.m21()}
	}
	if i == 2 {
		return Vec3[T]{
			lhs.m02(),
			lhs.m12(),
			lhs.m22(),
		}
	}

	panic(i)
}

func (lhs Mat3[T]) ToWGPU() [12]float32 {
	return [12]float32{
		float32(lhs.m00()), float32(lhs.m01()), float32(lhs.m02()), 0,
		float32(lhs.m10()), float32(lhs.m11()), float32(lhs.m12()), 0,
		float32(lhs.m20()), float32(lhs.m21()), float32(lhs.m22()), 0,
	}
}

func (lhs *Mat3[T]) m00() T {
	return lhs.values[0][0] + 1
}

func (lhs *Mat3[T]) m01() T {
	return lhs.values[0][1]
}

func (lhs *Mat3[T]) m02() T {
	return lhs.values[0][2]
}

func (lhs *Mat3[T]) m10() T {
	return lhs.values[1][0]
}

func (lhs *Mat3[T]) m11() T {
	return lhs.values[1][1] + 1
}

func (lhs *Mat3[T]) m12() T {
	return lhs.values[1][2]
}

func (lhs *Mat3[T]) m20() T {
	return lhs.values[2][0]
}

func (lhs *Mat3[T]) m21() T {
	return lhs.values[2][1]
}

func (lhs *Mat3[T]) m22() T {
	return lhs.values[2][2] + 1
}
