package orion

import (
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse/commands"
)

type DebugTextOptions struct {
	Transform   glm.Mat3f
	ColorScale  ColorScale
	ShadowColor Color
	TabWidth    uint
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

	textCommand.DrawText(dest.texture, commands.DrawDebugTextOptions{
		Text:        text,
		Transform:   opts.Transform,
		TextColor:   opts.ColorScale.ToColor(),
		ShadowColor: opts.ShadowColor,
		TabWidth:    tabWidth,
	})
}
