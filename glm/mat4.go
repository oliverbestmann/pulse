package glm

func TranslationMat4[T Numeric](x, y, z T) Mat4[T] {
	return Mat4[T]{
		values: [4][4]T{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{x, y, z, 0},
		},
	}
}

func ScaleMat4[T Numeric](x, y, z T) Mat4[T] {
	x -= 1
	y -= 1
	z -= 1

	return Mat4[T]{
		values: [4][4]T{
			{x, 0, 0, 0},
			{0, y, 0, 0},
			{0, 0, z, 0},
			{0, 0, 0, 0},
		},
	}
}

func RotationZMat4[T Numeric](angle Rad) Mat4[T] {
	fs, fc := fastSincos(angle)
	s := T(fs)
	c := T(fc) - 1

	return Mat4[T]{
		values: [4][4]T{
			{c, s, 0, 0},
			{-s, c, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
	}
}

func RotationXMat4[T Numeric](angle Rad) Mat4[T] {
	fs, fc := fastSincos(angle)
	s := T(fs)
	c := T(fc) - 1

	return Mat4[T]{
		values: [4][4]T{
			{0, 0, 0, 0},
			{0, c, s, 0},
			{0, -s, c, 0},
			{0, 0, 0, 0},
		},
	}
}

func RotationYMat4[T Numeric](angle Rad) Mat4[T] {
	fs, fc := fastSincos(angle)
	s := T(fs)
	c := T(fc) - 1

	return Mat4[T]{
		values: [4][4]T{
			{c, 0, s, 0},
			{0, 0, 0, 0},
			{-s, 0, c, 0},
			{0, 0, 0, 0},
		},
	}
}

func (m Mat4[T]) Translate(x, y, z T) Mat4[T] {
	return m.Mul(TranslationMat4[T](x, y, z))
}

func (m Mat4[T]) Scale(x, y, z T) Mat4[T] {
	return m.Mul(ScaleMat4[T](x, y, z))
}

func (m Mat4[T]) RotateX(angle Rad) Mat4[T] {
	return m.Mul(RotationXMat4[T](angle))
}

func (m Mat4[T]) RotateY(angle Rad) Mat4[T] {
	return m.Mul(RotationYMat4[T](angle))
}

func (m Mat4[T]) RotateZ(angle Rad) Mat4[T] {
	return m.Mul(RotationZMat4[T](angle))
}

func (m Mat4[T]) Transform(vec Vec4[T]) Vec4[T] {
	return Vec4[T]{
		(m.m00())*vec[0] + (m.m10())*vec[1] + (m.m20())*vec[2] + (m.m30())*vec[3],
		(m.m01())*vec[0] + (m.m11())*vec[1] + (m.m21())*vec[2] + (m.m31())*vec[3],
		(m.m02())*vec[0] + (m.m12())*vec[1] + (m.m22())*vec[2] + (m.m32())*vec[3],
		(m.m03())*vec[0] + (m.m13())*vec[1] + (m.m23())*vec[2] + (m.m33())*vec[3],
	}
}

func (m Mat4[T]) Transform3(vec Vec3[T]) Vec3[T] {
	return Vec3[T]{
		(m.m00())*vec[0] + (m.m10())*vec[1] + (m.m20() * vec[2]) + m.m30(),
		(m.m01())*vec[0] + (m.m11())*vec[1] + (m.m21() * vec[2]) + m.m31(),
		(m.m02())*vec[0] + (m.m12())*vec[1] + (m.m22() * vec[2]) + m.m32(),
	}
}

func (m Mat4[T]) Transform2(vec Vec2[T]) Vec2[T] {
	return Vec2[T]{
		(m.m00())*vec[0] + (m.m10())*vec[1] + m.m30(),
		(m.m01())*vec[0] + (m.m11())*vec[1] + m.m31(),
	}
}
