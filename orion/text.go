package orion

import (
	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/pulse/pulse/commands"
)

type DebugTextOptions struct {
	Transform   glm.Mat3f
	ColorScale  Color
	ShadowColor *Color
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

	shadowColor := pulse.ColorLinearRGBA(0, 0, 0, 0.5)
	if opts.ShadowColor != nil {
		shadowColor = *opts.ShadowColor
	}

	textCommand.DrawText(dest.texture, commands.DrawDebugTextOptions{
		Text:        text,
		Transform:   opts.Transform,
		TextColor:   opts.ColorScale,
		ShadowColor: shadowColor,
		TabWidth:    tabWidth,
	})
}
