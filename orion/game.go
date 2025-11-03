package orion

import "github.com/oliverbestmann/go3d/pulse"

type RenderTarget = pulse.RenderTarget

type Game interface {
	Initialize() error
	Update() error
	Draw(screen *Image)
}
