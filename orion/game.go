package orion

import (
	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/pulse"
)

type RenderTarget = pulse.RenderTarget

type Game interface {
	Layout(surfaceWidth, surfaceHeight uint32) LayoutOptions

	Initialize() error
	Update() error
	Draw(screen *Image)
}

type GameWithFinalizeDrawScreen interface {
	Game

	FinalizeDrawScreen(surface, offscreen *Image)
}

type LayoutOptions struct {
	// Defaults to the size of the surface
	Width  uint32
	Height uint32

	// Enable MSAA on the offscreen render target
	MSAA bool

	// Format of the offscreen buffer, if not specified, the default
	// texture format will be wgpu.TextureFormatBGRA8Unorm
	Format wgpu.TextureFormat
}

func (o LayoutOptions) withDefaults(surfaceWidth, surfaceHeight uint32) LayoutOptions {
	if o.Width == 0 {
		o.Width = surfaceWidth
	}

	if o.Height == 0 {
		o.Height = surfaceHeight
	}

	if o.Format == 0 {
		o.Format = wgpu.TextureFormatRGBA8Unorm
	}

	return o
}
