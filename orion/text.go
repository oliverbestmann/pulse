package orion

import (
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse"
)

type DebugTextOptions struct {
	Transform  glm.Mat3f
	ColorScale ColorScale
}

func DebugText(dest *Image, text string, opts *DebugTextOptions) {
	if opts == nil {
		opts = &DebugTextOptions{}
	}

	// switch to sprite rendering, as text rendering is just using the
	// sprite pipeline for now
	switchToCommand(spriteCommand.Get())

	err := textCommand.Get().DrawText(dest.renderTarget, pulse.DrawTextOptions{
		Text:      text,
		Transform: opts.Transform,
		Color:     opts.ColorScale.ToColor(),
	})
	handle(err, "render text")
}
