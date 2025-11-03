package orion

import (
	"github.com/oliverbestmann/go3d/pulse"
)

var clearCommand global[*pulse.ClearCommand]
var spriteCommand global[*pulse.SpriteCommand]
var mesh2dCommand global[*pulse.Mesh2dCommand]

func initializeCommands(ctx *pulse.Context) {
	clearCommand.set(pulse.NewClear(ctx))

	sprite, err := pulse.NewSpriteCommand(ctx)
	handle(err, "initialize sprite commands")
	spriteCommand.set(sprite)

	mesh2d, err := pulse.NewMesh2dCommands(ctx)
	handle(err, "initialize mesh2d commands")
	mesh2dCommand.set(mesh2d)
}

type command interface {
	Flush() error
}

var previousCommand command

func switchToCommand(next command) {
	flushCommand()
	previousCommand = next
}

func flushCommand() {
	defer func() { previousCommand = nil }()

	if previousCommand != nil {
		err := previousCommand.Flush()
		handle(err, "flush pending commands")
	}
}
