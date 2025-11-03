package glimpse

import "github.com/cogentcore/webgpu/wgpu"

type Window interface {
	GetSize() (uint32, uint32)
	SurfaceDescriptor() *wgpu.SurfaceDescriptor
	Run(render func() error) error
	Terminate()
}
