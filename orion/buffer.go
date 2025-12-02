package orion

import (
	"unsafe"

	"github.com/oliverbestmann/webgpu/wgpu"
)

type BufferInitDescriptor[T any] struct {
	Label    string
	Contents []T
	Usage    wgpu.BufferUsage
}

func CreateBufferInit[T any](desc BufferInitDescriptor[T]) *wgpu.Buffer {
	ctx := CurrentContext()

	return ctx.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label:    desc.Label,
		Usage:    desc.Usage,
		Contents: wgpu.ToBytes(desc.Contents),
	})
}

func WriteSliceToBuffer[T any](target *wgpu.Buffer, values []T) {
	CurrentContext().WriteBuffer(target, 0, wgpu.ToBytes(values))
}

func WriteValueToBuffer[T any](target *wgpu.Buffer, value T) {
	values := unsafe.Slice(&value, 1)
	WriteSliceToBuffer(target, values)
}
