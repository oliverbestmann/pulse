package orion

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/pulse"
)

type LoopState struct {
	Window glimpse.Window
	Game   Game
	Width  uint32
	Height uint32
}

func loopOnce(viewState *pulse.View, loopState *LoopState) {
	newWidth, newHeight := loopState.Window.GetSize()
	if loopState.Width != newWidth || loopState.Height != newHeight {
		loopState.Width = newWidth
		loopState.Height = newHeight

		slog.Info("Resize screen",
			slog.Int("width", int(newWidth)),
			slog.Int("height", int(newHeight)),
		)

		if err := viewState.Configure(newWidth, newHeight); err != nil {
			panic(fmt.Errorf("SizeChanged: %w", err))
		}
	}

	err := render(viewState, loopState.Game)
	if err != nil {
		fmt.Println("error occurred while rendering:", err)

		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "Surface timed out"):
		case strings.Contains(errStr, "Surface is outdated"):
		case strings.Contains(errStr, "Surface was lost"):
		default:
			panic(err)
		}
	}
}

func render(ctx *pulse.View, game Game) error {
	screen, err := ctx.Surface.GetCurrentTexture()
	if err != nil {
		panic(err)
	}

	screenGuard := pulse.NewReleaseGuard(screen)
	defer screenGuard.Release()

	screenView, err := screen.CreateView(nil)
	if err != nil {
		return fmt.Errorf("get texture: %w", err)
	}

	defer screenView.Release()

	screenTexture := ctx.AsTexture(screen, screenView)
	game.Draw(asImage(screenTexture))

	flushCommand()

	// present the rendered image
	ctx.Surface.Present()

	// no need to free the screen anymore
	screenGuard.Keep()

	return nil
}
