package orion

import (
	"github.com/oliverbestmann/go3d/pulse"
)

var clearCommand global[*pulse.ClearCommand]
var spriteCommand global[*pulse.SpriteCommand]
var mesh2dCommand global[*pulse.Mesh2dCommand]
var textCommand global[*pulse.TextCommand]

func initializeCommands(ctx *pulse.Context) {
	clearCommand.set(pulse.NewClear(ctx))

	sprite, err := pulse.NewSpriteCommand(ctx)
	handle(err, "initialize sprite command")
	spriteCommand.set(sprite)

	mesh2d, err := pulse.NewMesh2dCommand(ctx)
	handle(err, "initialize mesh2d command")
	mesh2dCommand.set(mesh2d)

	text, err := pulse.NewTextCommand(ctx, sprite)
	handle(err, "initialize text command")
	textCommand.set(text)
}

type command interface {
	Flush() error
}

var previousCommand command

func switchToCommand(next command) {
	if previousCommand != next {
		flushCommand()
	}

	previousCommand = next
}

func flushCommand() {
	defer func() { previousCommand = nil }()

	if previousCommand != nil {
		err := previousCommand.Flush()
		handle(err, "flush pending commands")
	}
}
