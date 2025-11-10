package orion

import (
	"github.com/oliverbestmann/go3d/pulse"
	"github.com/oliverbestmann/go3d/pulse/commands"
)

var clearCommand global[*commands.ClearCommand]
var spriteCommand global[*commands.SpriteCommand]
var mesh2dCommand global[*commands.Mesh2dCommand]
var textCommand global[*commands.TextCommand]

func initializeCommands(ctx *pulse.Context) {

	sprite, err := commands.NewSpriteCommand(ctx)
	Handle(err, "initialize sprite command")
	spriteCommand.set(sprite)

	clearCommand.set(commands.NewClear(ctx, sprite))

	mesh2d, err := commands.NewMesh2dCommand(ctx)
	Handle(err, "initialize mesh2d command")
	mesh2dCommand.set(mesh2d)

	text, err := commands.NewTextCommand(ctx, sprite)
	Handle(err, "initialize text command")
	textCommand.set(text)
}

type Command interface {
	Flush() error
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

		err := currentCommand.Flush()
		Handle(err, "flush pending commands")
	}
}
