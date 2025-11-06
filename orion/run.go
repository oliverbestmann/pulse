package orion

import (
	"errors"
	"fmt"

	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/pulse"
)

type RunGameOptions struct {
	// game to run. This is the only field that is required
	Game Game

	WindowWidth  int
	WindowHeight int
	WindowTitle  string
}

func RunGame(opts RunGameOptions) error {
	game := opts.Game
	if game == nil {
		return errors.New("Game must not be nil")
	}

	if opts.WindowWidth == 0 {
		opts.WindowWidth = 1000
	}

	if opts.WindowHeight == 0 {
		opts.WindowHeight = 600
	}

	if opts.WindowTitle == "" {
		opts.WindowTitle = "Orion"
	}

	// create a new window (or canvas)
	win, err := glimpse.NewWindow(
		opts.WindowWidth,
		opts.WindowHeight,
		opts.WindowTitle,
	)
	if err != nil {
		return fmt.Errorf("create window: %w", err)
	}

	defer win.Terminate()

	// initialize the webgpu device
	ctx, err := pulse.New(win.SurfaceDescriptor())
	if err != nil {
		return fmt.Errorf("initializing wgpu: %w", err)
	}

	defer ctx.Release()

	// initialize the view
	view, err := pulse.NewView(ctx, false)
	if err != nil {
		return fmt.Errorf("create view: %w", err)
	}

	defer view.Release()

	currentWindow.set(win)
	currentContext.set(ctx)
	currentView.set(view)

	initializeCommands(ctx)

	loopState := &LoopState{
		Window: win,
		Game:   game,
	}

	return win.Run(func(inputState glimpse.UpdateInputState) error {
		// do the actual rendering here
		return loopOnce(view, loopState, inputState)
	})
}
