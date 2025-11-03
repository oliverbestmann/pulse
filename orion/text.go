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

	err := textCommand.Get().DrawText(dest.renderTarget, pulse.DrawTextOptions{
		Text:      text,
		Transform: opts.Transform,
		Color:     opts.ColorScale.ToColor(),
	})
	handle(err, "render text")
}
