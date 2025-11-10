package orion

import (
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse/commands"
)

type DebugTextOptions struct {
	Transform  glm.Mat3f
	ColorScale ColorScale
	TabWidth   uint
}

func DebugText(dest *Image, text string, opts *DebugTextOptions) {
	if opts == nil {
		opts = &DebugTextOptions{}
	}

	var tabWidth uint = 8
	if opts.TabWidth > 0 {
		tabWidth = opts.TabWidth
	}

	textCommand := textCommand.Get()
	SwitchToCommand(textCommand)

	err := textCommand.DrawText(dest.texture, commands.DrawTextOptions{
		Text:      text,
		Transform: opts.Transform,
		Color:     opts.ColorScale.ToColor(),
		TabWidth:  tabWidth,
	})
	Handle(err, "render text")
}
