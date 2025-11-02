package orion

import "github.com/oliverbestmann/go3d/pulse"

type Game interface {
	Init() error
	Draw(screen *pulse.RenderTarget) error
}
