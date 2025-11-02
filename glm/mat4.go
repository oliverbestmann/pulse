package glm

type Mat4[T numeric] [16]T

func Mat4FromQuaternion[T numeric](quat Quaternion[T]) Mat4[T] {
	x2 := quat.V[0] + quat.V[0]
	y2 := quat.V[1] + quat.V[1]
	z2 := quat.V[2] + quat.V[2]

	xx2 := x2 * quat.V[0]
	xy2 := x2 * quat.V[1]
	xz2 := x2 * quat.V[2]

	yy2 := y2 * quat.V[1]
	yz2 := y2 * quat.V[2]
	zz2 := z2 * quat.V[2]

	sy2 := y2 * quat.S
	sz2 := z2 * quat.S
	sx2 := x2 * quat.S

	return Mat4[T]{
		1 - yy2 - zz2, xy2 + sz2, xz2 - sy2, 0,
		xy2 - sz2, 1 - xx2 - zz2, yz2 + sx2, 0,
		xz2 + sy2, yz2 - sx2, 1 - xx2 - yy2, 0,
		0, 0, 0, 1,
	}
}

func IdentityMat4[T numeric]() Mat4[T] {
	return Mat4[T]{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func TranslationMat4[T numeric](x, y, z T) Mat4[T] {
	return Mat4[T]{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func RotationZMat4[T numeric](angle Rad) Mat4[T] {
	fs, fc := fastSincos(angle)
	s := T(fs)
	c := T(fc)

	return Mat4[T]{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func RotationXMat4[T numeric](angle Rad) Mat4[T] {
	fs, fc := fastSincos(angle)
	s := T(fs)
	c := T(fc)

	return Mat4[T]{
		0, 0, 1, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

func RotationYMat4[T numeric](angle Rad) Mat4[T] {
	fs, fc := fastSincos(angle)
	s := T(fs)
	c := T(fc)

	return Mat4[T]{
		c, 0, s, 0,
		0, 0, 0, 0,
		-s, 0, c, 0,
		0, 0, 0, 1,
	}
}

func ScaleMat4[T numeric](x, y, z T) Mat4[T] {
	return Mat4[T]{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func (lhs Mat4[T]) RotateX(angle Rad) Mat4[T] {
	return lhs.Mul(RotationXMat4[T](angle))
}

func (lhs Mat4[T]) RotateY(angle Rad) Mat4[T] {
	return lhs.Mul(RotationYMat4[T](angle))
}

func (lhs Mat4[T]) RotateZ(angle Rad) Mat4[T] {
	return lhs.Mul(RotationZMat4[T](angle))
}

func (lhs Mat4[T]) Scale(x, y, z T) Mat4[T] {
	return lhs.Mul(ScaleMat4[T](x, y, z))
}

func (lhs Mat4[T]) Translate(x, y, z T) Mat4[T] {
	return lhs.Mul(TranslationMat4[T](x, y, z))
}

func (lhs Mat4[T]) IsZero() bool {
	return lhs == Mat4[T]{}
}

func (lhs Mat4[T]) Mul(rhs Mat4[T]) Mat4[T] {
	return Mat4[T]{
		lhs[0]*rhs[0] + lhs[4]*rhs[1] + lhs[8]*rhs[2] + lhs[12]*rhs[3],
		lhs[1]*rhs[0] + lhs[5]*rhs[1] + lhs[9]*rhs[2] + lhs[13]*rhs[3],
		lhs[2]*rhs[0] + lhs[6]*rhs[1] + lhs[10]*rhs[2] + lhs[14]*rhs[3],
		lhs[3]*rhs[0] + lhs[7]*rhs[1] + lhs[11]*rhs[2] + lhs[15]*rhs[3],
		lhs[0]*rhs[4] + lhs[4]*rhs[5] + lhs[8]*rhs[6] + lhs[12]*rhs[7],
		lhs[1]*rhs[4] + lhs[5]*rhs[5] + lhs[9]*rhs[6] + lhs[13]*rhs[7],
		lhs[2]*rhs[4] + lhs[6]*rhs[5] + lhs[10]*rhs[6] + lhs[14]*rhs[7],
		lhs[3]*rhs[4] + lhs[7]*rhs[5] + lhs[11]*rhs[6] + lhs[15]*rhs[7],
		lhs[0]*rhs[8] + lhs[4]*rhs[9] + lhs[8]*rhs[10] + lhs[12]*rhs[11],
		lhs[1]*rhs[8] + lhs[5]*rhs[9] + lhs[9]*rhs[10] + lhs[13]*rhs[11],
		lhs[2]*rhs[8] + lhs[6]*rhs[9] + lhs[10]*rhs[10] + lhs[14]*rhs[11],
		lhs[3]*rhs[8] + lhs[7]*rhs[9] + lhs[11]*rhs[10] + lhs[15]*rhs[11],
		lhs[0]*rhs[12] + lhs[4]*rhs[13] + lhs[8]*rhs[14] + lhs[12]*rhs[15],
		lhs[1]*rhs[12] + lhs[5]*rhs[13] + lhs[9]*rhs[14] + lhs[13]*rhs[15],
		lhs[2]*rhs[12] + lhs[6]*rhs[13] + lhs[10]*rhs[14] + lhs[14]*rhs[15],
		lhs[3]*rhs[12] + lhs[7]*rhs[13] + lhs[11]*rhs[14] + lhs[15]*rhs[15],
	}
}

func (lhs Mat4[T]) Transform(rhs Vec4[T]) Vec4[T] {
	return Vec4[T]{
		lhs[0]*rhs[0] + lhs[4]*rhs[1] + lhs[8]*rhs[2] + lhs[12]*rhs[3],
		lhs[1]*rhs[0] + lhs[5]*rhs[1] + lhs[9]*rhs[2] + lhs[13]*rhs[3],
		lhs[2]*rhs[0] + lhs[6]*rhs[1] + lhs[10]*rhs[2] + lhs[14]*rhs[3],
		lhs[3]*rhs[0] + lhs[7]*rhs[1] + lhs[11]*rhs[2] + lhs[15]*rhs[3],
	}
}
