package orion

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse"
)

type LoopState struct {
	Window        glimpse.Window
	Game          Game
	SurfaceWidth  uint32
	SurfaceHeight uint32
	Initialized   bool

	Canvas *Image
}

func loopOnce(viewState *pulse.View, loopState *LoopState, inputState glimpse.UpdateInputState) error {
	// get surface size for next frame
	surfaceWidth, surfaceHeight := loopState.Window.GetSize()

	// reconfigure surface if needed
	if loopState.SurfaceWidth != surfaceWidth || loopState.SurfaceHeight != surfaceHeight {
		slog.Debug("Resize surface",
			slog.Int("width", int(surfaceWidth)),
			slog.Int("height", int(surfaceHeight)),
		)

		if err := viewState.Configure(surfaceWidth, surfaceHeight); err != nil {
			return fmt.Errorf("resize surface: %w", err)
		}

		loopState.SurfaceWidth = surfaceWidth
		loopState.SurfaceHeight = surfaceHeight
	}

	// get requested layout
	layout := loopState.Game.
		Layout(surfaceWidth, surfaceHeight).
		withDefaults(surfaceWidth, surfaceHeight)

	// request a new surface if needed
	if !layoutIsCompatible(layout, loopState.Canvas) {
		loopState.Canvas = NewImage(layout.Width, layout.Height, &NewImageOptions{
			Label:  "OffscreenTarget",
			Format: layout.Format,
			MSAA:   layout.MSAA,
		})
	}

	screen, err := CurrentContext().Surface.GetCurrentTexture()
	if err != nil {
		return fmt.Errorf("get current texture: %w", err)
	}

	// get input after waiting for a texture to keep input lag low
	currentInputState.reset()
	currentInputState.set(inputState())

	if !loopState.Initialized {
		loopState.Initialized = true

		if err := loopState.Game.Initialize(); err != nil {
			return fmt.Errorf("initialize game: %w", err)
		}
	}

	if err := loopState.Game.Update(); err != nil {
		return fmt.Errorf("update game: %w", err)
	}

	// draw to canvas first
	loopState.Game.Draw(loopState.Canvas)

	// finalize drawing
	err = present(viewState, screen, loopState.Canvas, finalizeDrawScreenOf(loopState.Game))
	if err != nil {
		fmt.Println("error occurred while rendering:", err)

		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "Surface timed out"):
		case strings.Contains(errStr, "Surface is outdated"):
		case strings.Contains(errStr, "Surface was lost"):
		default:
			return fmt.Errorf("render frame: %w", err)
		}
	}

	return nil
}

func present(ctx *pulse.View, screen *wgpu.Texture, canvas *Image, draw FinalizeDrawScreen) error {
	screenGuard := pulse.NewReleaseGuard(screen)
	defer screenGuard.Release()

	screenView, err := screen.CreateView(nil)
	if err != nil {
		return fmt.Errorf("get texture: %w", err)
	}

	defer screenView.Release()

	surfaceTexture := ctx.AsTexture(screen, screenView)
	surfaceImage := asImage(surfaceTexture)

	// then paint canvas to the screen surface
	draw(surfaceImage, canvas)

	// flushes any outstanding pipelines
	SwitchToCommand(nil)

	// present the rendered image
	ctx.Surface.Present()

	// no need to free the screen anymore
	screenGuard.Keep()

	return nil
}

func layoutIsCompatible(layout LayoutOptions, canvas *Image) bool {
	return canvas != nil &&
		canvas.Width() == layout.Width &&
		canvas.Height() == layout.Height &&
		canvas.Format() == layout.Format &&
		canvas.MSAA() == layout.MSAA
}

type FinalizeDrawScreen = func(screen, canvas *Image)

func finalizeDrawScreenOf(game Game) FinalizeDrawScreen {
	if game, ok := game.(GameWithFinalizeDrawScreen); ok {
		return game.FinalizeDrawScreen
	}

	return DefaultFinalizeDrawScreen
}

func DefaultFinalizeDrawScreen(screen, canvas *Image) {
	cw, ch := canvas.Size().XY()
	sw, sh := screen.Size().XY()

	canvasAspect := cw / ch
	screenAspect := sw / sh

	var scale float32
	var xOffset, yOffset float32 = 1, 1

	if canvasAspect >= screenAspect {
		scale = sw / cw
		yOffset = (sh - ch*scale) / 2
	} else {
		scale = sh / ch
		xOffset = (sw - cw*scale) / 2
	}

	tr := glm.TranslationMat3(xOffset, yOffset).Scale(scale, scale)

	screen.DrawImage(canvas, &DrawImageOptions{
		Transform:  tr,
		FilterMode: wgpu.FilterModeLinear,
		BlendState: wgpu.BlendStateReplace,
	})
}
