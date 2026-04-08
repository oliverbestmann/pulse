package orion

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/oliverbestmann/pulse/vyn"
	"github.com/oliverbestmann/pulse/wx"
)

type RunGameOptions struct {
	// game to run. This is the only field that is required
	Game Game

	WindowWidth     int
	WindowHeight    int
	WindowTitle     string
	WindowResizable bool
}

func RunGame(opts RunGameOptions) error {
	game := opts.Game

	if game == nil {
		return errors.New("game must not be nil")
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
	win, err := vyn.NewWindow(
		opts.WindowWidth,
		opts.WindowHeight,
		opts.WindowTitle,
		opts.WindowResizable,
	)
	if err != nil {
		return fmt.Errorf("create window: %w", err)
	}

	defer win.Terminate()

	// initialize the webgpu device
	ctx, err := wx.New(win.SurfaceDescriptor())
	if err != nil {
		return fmt.Errorf("initializing wgpu: %w", err)
	}

	defer ctx.Release()

	if runtime.GOOS != "js" {
		// print adapter info
		adapterInfo := ctx.Adapter.GetInfo()
		fmt.Printf("Using device: %s\n", adapterInfo.Device)
		fmt.Printf("Description:  %s\n", adapterInfo.Description)
		fmt.Printf("Backend:      %s\n", adapterInfo.BackendType)
		fmt.Printf("Vendor:       %s\n", adapterInfo.Vendor)
	}

	// initialize the view
	view := wx.NewSurface(ctx, false, false)

	defer view.Release()

	currentWindow.set(win)
	currentContext.set(ctx)
	currentView.set(view)

	initializeCommands(ctx)

	loopState := &LoopState{
		Window: win,
		Game:   game,
	}

	err = win.Run(func(inputState vyn.UpdateInputState) error {
		// do the actual rendering here
		return loopOnce(view, loopState, inputState)
	})

	if errors.Is(err, ExitApp) {
		return nil
	}

	return err
}
