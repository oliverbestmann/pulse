package orion

import (
	"runtime"

	"github.com/cogentcore/webgpu/wgpu"
)

func CreateBuffer(desc wgpu.BufferDescriptor) *wgpu.Buffer {
	ctx := CurrentContext()

	// allocate a buffer to write to
	buf, err := ctx.CreateBuffer(&desc)
	handle(err, "create buffer label=%q", desc.Label)

	runtime.SetFinalizer(buf, func(buf *wgpu.Buffer) {
		buf.Release()
	})

	return buf
}

func CreateBufferInit(desc wgpu.BufferInitDescriptor) *wgpu.Buffer {
	ctx := CurrentContext()

	// allocate a buffer to write to
	buf, err := ctx.CreateBufferInit(&desc)
	handle(err, "create and init buffer label=%q", desc.Label)

	runtime.SetFinalizer(buf, func(buf *wgpu.Buffer) {
		buf.Release()
	})

	return buf
}

func WriteToBuffer[T any](target *wgpu.Buffer, values []T) {
	ctx := CurrentContext()

	// copy CPU to GPU
	queue := ctx.GetQueue()
	defer queue.Release()

	err := queue.WriteBuffer(target, 0, wgpu.ToBytes(values))
	handle(err, "write to buffer")
}
