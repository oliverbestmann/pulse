package orion

import (
	"errors"

	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/webgpu/wgpu"
)

// ExitApp can be returned from Game.Update to exit the app
var ExitApp = errors.New("app is existing")

type Game interface {
	Layout(surfaceWidth, surfaceHeight uint32) LayoutOptions

	Initialize() error
	Update() error
	Draw(screen *Image)

	// DrawToSurface draws the offscreen texture to the actual window surface.
	DrawToSurface(surface, offscreen *Image)
}

type LayoutOptions struct {
	// Defaults to the size of the surface
	Width  uint32
	Height uint32

	// Format of the offscreen buffer, if not specified, the default
	// texture format will be wgpu.TextureFormatBGRA8Unorm
	Format wgpu.TextureFormat

	// Enable MSAA on the offscreen render target
	MSAA bool
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

// DefaultGame implements a simple game that does nothing. You can embed it into
// your own game struct to add default implementations to satisfy the Game interface.
type DefaultGame struct{}

func (d DefaultGame) Layout(surfaceWidth, surfaceHeight uint32) LayoutOptions {
	return LayoutOptions{
		Width:  surfaceWidth,
		Height: surfaceHeight,
		Format: wgpu.TextureFormatRGBA8Unorm,
		MSAA:   false,
	}
}

func (d DefaultGame) Initialize() error {
	return nil
}

func (d DefaultGame) Update() error {
	return nil
}

func (d DefaultGame) Draw(screen *Image) {
	screen.Clear(pulse.ColorLinearRGBA(0.7, 0.7, 0.8, 1.0))
}

func (d DefaultGame) DrawToSurface(surface, offscreen *Image) {
	DefaultDrawToSurface(surface, offscreen, wgpu.FilterModeLinear)
}
