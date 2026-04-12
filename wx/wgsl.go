package wx

import (
	"unsafe"

	"github.com/oliverbestmann/pulse/glm"
)

type StructWriter struct {
	buf []byte
}

func (s *StructWriter) Clear() {
	s.buf = s.buf[:0]
}

func (s *StructWriter) Bytes() []byte {
	return s.buf
}

func (s *StructWriter) AppendFloat32(value float32) {
	s.buf = appendTo(s.buf, value, 4, 4)
}

func (s *StructWriter) AppendInt(value int32) {
	s.buf = appendTo(s.buf, value, 4, 4)
}

func (s *StructWriter) AppendUint(value uint32) {
	s.buf = appendTo(s.buf, value, 4, 4)
}

func (s *StructWriter) AppendVec2f(value glm.Vec2f) {
	s.buf = appendTo(s.buf, value, 8, 8)
}

func (s *StructWriter) AppendVec3f(value glm.Vec4f) {
	s.buf = appendTo(s.buf, value, 16, 12)
}

func (s *StructWriter) AppendVec4f(value glm.Vec4f) {
	s.buf = appendTo(s.buf, value, 16, 16)
}

func (s *StructWriter) AppendMat2f(value glm.Mat4f) {
	s.buf = appendTo(s.buf, value.Components(), 8, 16)
}

func (s *StructWriter) AppendMat3f(value glm.Mat3f) {
	values := value.Components()
	s.buf = appendTo(s.buf, values[0], 16, 16)
	s.buf = appendTo(s.buf, values[1], 16, 16)
	s.buf = appendTo(s.buf, values[2], 16, 16)
}

func (s *StructWriter) AppendMat4f(value glm.Mat4f) {
	s.buf = appendTo(s.buf, value.Components(), 16, 4*16)
}

type InstanceWriter struct {
	buf []byte
}

func (s *InstanceWriter) Clear() {
	s.buf = s.buf[:0]
}

func (s *InstanceWriter) Len() int {
	return len(s.buf)
}

func (s *InstanceWriter) Bytes() []byte {
	return s.buf
}

func (s *InstanceWriter) AppendFloat32(value float32) {
	s.buf = appendTo(s.buf, value, 1, 4)
}

func (s *InstanceWriter) AppendInt(value int32) {
	s.buf = appendTo(s.buf, value, 1, 4)
}

func (s *InstanceWriter) AppendUint(value uint32) {
	s.buf = appendTo(s.buf, value, 1, 4)
}

func (s *InstanceWriter) AppendVec2f(value glm.Vec2f) {
	s.buf = appendTo(s.buf, value, 1, 8)
}

func (s *InstanceWriter) AppendVec3f(value glm.Vec4f) {
	s.buf = appendTo(s.buf, value, 1, 12)
}

func (s *InstanceWriter) AppendVec4f(value glm.Vec4f) {
	s.buf = appendTo(s.buf, value, 1, 16)
}

func appendTo[T any](buf []byte, value T, align, size int) []byte {
	if unsafe.Sizeof(value) > uintptr(size) {
		panic("value is larger than its expected size")
	}

	for len(buf)%align != 0 {
		buf = append(buf, 0)
	}

	ptrValue := (*byte)(unsafe.Pointer(&value))
	bufValue := unsafe.Slice(ptrValue, unsafe.Sizeof(value))
	buf = append(buf, bufValue...)

	// add padding
	padding := int(unsafe.Sizeof(value)) - size
	if padding > 0 {
		buf = append(buf, make([]byte, padding)...)
	}

	return buf
}
