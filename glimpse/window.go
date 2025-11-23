package glimpse

import "github.com/oliverbestmann/webgpu/wgpu"

type Window interface {
	GetSize() (uint32, uint32)
	SurfaceDescriptor() *wgpu.SurfaceDescriptor
	Run(render func(inputState UpdateInputState) error) error
	Terminate()
}
