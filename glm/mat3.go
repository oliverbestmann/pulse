package glm

func TranslationMat3[T Numeric](x, y T) Mat3[T] {
	return Mat3Of([3][3]T{
		{1, 0, 0},
		{0, 1, 0},
		{x, y, 1},
	})
}

func RotationMat3[T Numeric](angle Rad) Mat3[T] {
	s, c := fastSincos(angle)

	return Mat3Of([3][3]T{
		{T(c), T(s), 0},
		{-T(s), T(c), 0},
		{0, 0, 1},
	})
}

func ScaleMat3[T Numeric](x, y T) Mat3[T] {
	return Mat3Of([3][3]T{
		{x, 0, 0},
		{0, y, 0},
		{0, 0, 1},
	})
}

func (m Mat3[T]) Translate(x, y T) Mat3[T] {
	return m.Mul(TranslationMat3[T](x, y))
}

func (m Mat3[T]) Rotate(angle Rad) Mat3[T] {
	return m.Mul(RotationMat3[T](angle))
}

func (m Mat3[T]) Scale(x, y T) Mat3[T] {
	return m.Mul(ScaleMat3[T](x, y))
}

func (m Mat3[T]) Row(i int) Vec3[T] {
	if i == 0 {
		return Vec3[T]{
			m.m00(),
			m.m10(),
			m.m20(),
		}
	}
	if i == 1 {
		return Vec3[T]{m.m01(), m.m11(), m.m21()}
	}
	if i == 2 {
		return Vec3[T]{
			m.m02(),
			m.m12(),
			m.m22(),
		}
	}

	panic(i)
}

func (m Mat3[T]) ToWGPU() [12]float32 {
	return [12]float32{
		float32(m.m00()), float32(m.m01()), float32(m.m02()), 0,
		float32(m.m10()), float32(m.m11()), float32(m.m12()), 0,
		float32(m.m20()), float32(m.m21()), float32(m.m22()), 0,
	}
}

func (m Mat3[T]) Invert() Mat3[T] {
	inv, ok := m.TryInvert()
	if !ok {
		panic("matrix not invertible")
	}

	return inv
}

func (m Mat3[T]) TryInvert() (Mat3[T], bool) {
	var inv [3][3]T

	// determinant
	det := m.m00()*(m.m11()*m.m22()-m.m12()*m.m21()) -
		m.m01()*(m.m10()*m.m22()-m.m12()*m.m20()) +
		m.m02()*(m.m10()*m.m21()-m.m11()*m.m20())

	if det == 0 {
		// singular
		return Mat3[T]{}, false
	}

	inv[0][0] = (m.m11()*m.m22() - m.m12()*m.m21()) / det
	inv[0][1] = (m.m02()*m.m21() - m.m01()*m.m22()) / det
	inv[0][2] = (m.m01()*m.m12() - m.m02()*m.m11()) / det

	inv[1][0] = (m.m12()*m.m20() - m.m10()*m.m22()) / det
	inv[1][1] = (m.m00()*m.m22() - m.m02()*m.m20()) / det
	inv[1][2] = (m.m02()*m.m10() - m.m00()*m.m12()) / det

	inv[2][0] = (m.m10()*m.m21() - m.m11()*m.m20()) / det
	inv[2][1] = (m.m01()*m.m20() - m.m00()*m.m21()) / det
	inv[2][2] = (m.m00()*m.m11() - m.m01()*m.m10()) / det

	return Mat3Of(inv), true
}

func (m Mat3[T]) Transform(vec Vec3[T]) Vec3[T] {
	return Vec3[T]{
		(m.m00())*vec[0] + (m.m10())*vec[1] + (m.m20())*vec[2],
		(m.m01())*vec[0] + (m.m11())*vec[1] + (m.m21())*vec[2],
		(m.m02())*vec[0] + (m.m12())*vec[1] + (m.m22())*vec[2],
	}
}

func (m Mat3[T]) Transform2(vec Vec2[T]) Vec2[T] {
	return Vec2[T]{
		(m.m00())*vec[0] + (m.m10())*vec[1] + (m.m20()),
		(m.m01())*vec[0] + (m.m11())*vec[1] + (m.m21()),
	}
}
