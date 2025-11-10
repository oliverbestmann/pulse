package orion

import (
	"github.com/cogentcore/webgpu/wgpu"
)

func CreateBuffer(desc wgpu.BufferDescriptor) *wgpu.Buffer {
	ctx := CurrentContext()

	// allocate a buffer to write to
	buf, err := ctx.CreateBuffer(&desc)
	Handle(err, "create buffer label=%q", desc.Label)

	return buf
}

func CreateBufferInit(desc wgpu.BufferInitDescriptor) *wgpu.Buffer {
	ctx := CurrentContext()

	// allocate a buffer to write to
	buf, err := ctx.CreateBufferInit(&desc)
	Handle(err, "create and init buffer label=%q", desc.Label)

	return buf
}

func WriteToBuffer[T any](target *wgpu.Buffer, values []T) {
	err := CurrentContext().WriteBuffer(target, 0, wgpu.ToBytes(values))
	Handle(err, "write to buffer")
}
