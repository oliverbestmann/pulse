package glm

type float interface {
	~float32 | ~float64
}

type numeric interface {
	float | uint32
}
