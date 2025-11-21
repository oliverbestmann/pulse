package orion

import (
	"fmt"
	"log/slog"

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
	DebugOverlay.StartFrame()

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
		if loopState.Canvas != nil {
			loopState.Canvas.Texture().Release()
			loopState.Canvas = nil
		}

		slog.Info("Allocate new offscreen render target",
			slog.Int("width", int(layout.Width)),
			slog.Int("height", int(layout.Height)))

		loopState.Canvas = NewImage(layout.Width, layout.Height, &NewImageOptions{
			Label:  "OffscreenTarget",
			Format: layout.Format,
			MSAA:   layout.MSAA,
		})
	}

	DebugOverlay.StartGetCurrentTexture()

	// get the surface texture (the actual screen)
	surface, err := CurrentContext().Surface.GetCurrentTexture()
	if err != nil {
		return fmt.Errorf("get current texture: %w", err)
	}

	defer func() {
		if surface != nil {
			surface.Release()
		}
	}()

	// get input after waiting for a texture to keep input lag low
	currentInputState.reset()
	currentInputState.set(inputState())

	// calculate screen transform to map input cursor/touch events
	updateScreenTransform(surface, loopState.Canvas.Sizef())

	// run game.Initialize and game.Update
	err = performGameUpdate(loopState)
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}

	// draw to canvas first
	DebugOverlay.StartGameDraw()
	loopState.Game.Draw(loopState.Canvas)

	// finalize drawing
	err = drawToSurface(viewState, loopState.Game, surface, loopState.Canvas)
	if err != nil {
		return fmt.Errorf("drawToSurface: %w", err)
	}

	// present the rendered image
	viewState.Surface.Present()

	// we do not need to release the screen if present was successful
	surface = nil

	DebugOverlay.EndFrame()

	return nil
}

func performGameUpdate(loopState *LoopState) error {
	DebugOverlay.StartGameUpdate()

	if !loopState.Initialized {
		loopState.Initialized = true

		if err := loopState.Game.Initialize(); err != nil {
			return fmt.Errorf("initialize game: %w", err)
		}
	}

	if err := loopState.Game.Update(); err != nil {
		return fmt.Errorf("update game: %w", err)
	}
	return nil
}

func updateScreenTransform(surface *wgpu.Texture, offscreenSize glm.Vec2f) {
	screenTransform := DefaultScreenTransform(
		glm.Vec2f{float32(surface.GetWidth()), float32(surface.GetHeight())},
		offscreenSize,
	)

	screenTransformInv := DefaultScreenTransformInv(
		glm.Vec2f{float32(surface.GetWidth()), float32(surface.GetHeight())},
		offscreenSize,
	)

	currentScreenTransform.reset()
	currentScreenTransform.set(screenTransform)

	currentScreenTransformInv.reset()
	currentScreenTransformInv.set(screenTransformInv)
}

func drawToSurface(ctx *pulse.View, game Game, surface *wgpu.Texture, screen *Image) error {
	surfaceView, err := surface.CreateView(nil)
	if err != nil {
		return fmt.Errorf("get texture: %w", err)
	}

	defer surfaceView.Release()

	surfaceTexture := ctx.SurfaceAsTexture(surface, surfaceView)
	surfaceImage := asImage(surfaceTexture)

	// then paint canvas to the surface

	game.DrawToSurface(surfaceImage, screen)

	// flushes any outstanding pipelines
	SwitchToCommand(nil)

	return nil
}

func layoutIsCompatible(layout LayoutOptions, canvas *Image) bool {
	return canvas != nil &&
		canvas.Width() == layout.Width &&
		canvas.Height() == layout.Height &&
		canvas.Format() == layout.Format &&
		canvas.MSAA() == layout.MSAA
}

func DefaultScreenTransform(surfaceSize, screenSize glm.Vec2f) glm.Mat3[float32] {
	cw, ch := screenSize.XY()
	sw, sh := surfaceSize.XY()

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

	return glm.TranslationMat3(xOffset, yOffset).Scale(scale, scale)
}

func DefaultScreenTransformInv(surfaceSize, screenSize glm.Vec2f) glm.Mat3[float32] {
	cw, ch := screenSize.XY()
	sw, sh := surfaceSize.XY()

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

	sm := glm.ScaleMat3(1.0/scale, 1.0/scale)
	return sm.Mul(glm.TranslationMat3(-xOffset, -yOffset))
}

func DefaultDrawToSurface(surface, offscreen *Image, filter wgpu.FilterMode) {
	surface.Clear(Color{0, 0, 0, 1})

	surface.DrawImage(offscreen, &DrawImageOptions{
		Transform:  DefaultScreenTransform(surface.Sizef(), offscreen.Sizef()),
		FilterMode: filter,
		BlendState: wgpu.BlendStateReplace,
	})
}
