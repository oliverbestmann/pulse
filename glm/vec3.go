package glm

func (lhs Vec3[T]) Cross(rhs Vec3[T]) Vec3[T] {
	return Vec3[T]{
		lhs[1]*rhs[2] - rhs[1]*lhs[2],
		lhs[2]*rhs[0] - rhs[2]*lhs[0],
		lhs[0]*rhs[1] - rhs[0]*lhs[1],
	}
}
