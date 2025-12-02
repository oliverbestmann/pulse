package orion

import (
	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/pulse/pulse/commands"
)

var clearCommand global[*commands.ClearCommand]
var spriteCommand global[*commands.SpriteCommand]
var mesh2dCommand global[*commands.Mesh2dCommand]
var textCommand global[*commands.DebugTextCommand]

func initializeCommands(ctx *pulse.Context) {
	sprite := commands.NewSpriteCommand(ctx)
	spriteCommand.set(sprite)

	clearCommand.set(commands.NewClear(ctx, sprite))

	mesh2d := commands.NewMesh2dCommand(ctx)
	mesh2dCommand.set(mesh2d)

	text := commands.NewDebugTextCommand(ctx, sprite)
	textCommand.set(text)
}

type Command interface {
	Flush()
}

var currentCommand Command

// SwitchToCommand flushes the current command and records `next`
// as the new current command.
func SwitchToCommand(next Command) {
	if currentCommand != next && currentCommand != nil {
		flushCurrentCommand()
	}

	currentCommand = next
}

func flushCurrentCommand() {
	if currentCommand != nil {
		defer func() { currentCommand = nil }()

		currentCommand.Flush()
	}
}
