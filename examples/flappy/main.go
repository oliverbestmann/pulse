package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	_ "image/png"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/orion"
)

//go:embed flappy.png
var _flappy_png []byte

type Game struct {
	orion.DefaultGame

	lastTime time.Time

	background *orion.Image
	foreground *orion.Image
	bird       *orion.Image
	pipe       *orion.Image

	xOffset   float32
	yOffset   float32
	yVelocity float32

	pipes []float32

	debugOverlay bool
}

func (g *Game) Layout(surfaceWidth, surfaceHeight uint32) orion.LayoutOptions {
	return orion.LayoutOptions{
		Width:  480,
		Height: 256,
	}
}

func (g *Game) Initialize() error {
	spriteSheet, err := orion.DecodeImageFromBytes(_flappy_png)
	if err != nil {
		return fmt.Errorf("load sprite sheet: %w", err)
	}

	g.background = spriteSheet.SubImage(0, 0, 50, 256)
	g.foreground = spriteSheet.SubImage(293, 0, 166, 56)
	g.pipe = spriteSheet.SubImage(84, 323, 26, 160)
	g.bird = spriteSheet.SubImage(0, 486, 24, 24)

	g.xOffset = 100
	g.yOffset = 128
	g.yVelocity = 0

	for range 20 {
		g.pipes = append(g.pipes, rand.Float32()*64)
	}

	return nil
}

func (g *Game) Update() error {
	now := time.Now()
	if g.lastTime.IsZero() {
		g.lastTime = now
	}

	dt := float32(now.Sub(g.lastTime).Seconds())
	g.lastTime = now

	g.xOffset += dt * 100

	g.yVelocity += 300 * dt
	g.yOffset += g.yVelocity * dt

	if orion.IsKeyJustPressed(glimpse.KeySpace) || orion.IsMouseButtonJustPressed(orion.MouseButton(0)) {
		g.yVelocity = -200
	}

	if orion.IsKeyJustPressed(glimpse.KeyD) {
		orion.DebugOverlay.RunGC = true
		orion.DebugOverlay.Toggle()
	}

	if orion.IsKeyJustPressed(glimpse.KeyEscape) {
		return orion.ExitApp
	}

	return nil
}

func (g *Game) Draw(screen *orion.Image) {
	// screen.Clear(orion.Color{0.5, 0.5, 0.8, 1.0})

	cam := glm.TranslationMat3(-g.xOffset, 0)

	g.drawTiles(screen, g.background, cam, 0, 0.5)

	modelTr := glm.TranslationMat3(g.xOffset+100, g.yOffset)
	screen.DrawImage(g.bird, &orion.DrawImageOptions{Transform: cam.Mul(modelTr)})

	// draw the pipes
	for idx, pipeY := range g.pipes {
		pipeX := float32(idx*256 + 400)

		pipeTr := cam.Translate(pipeX, 32+pipeY).Scale(1, -1)
		screen.DrawImage(g.pipe, &orion.DrawImageOptions{Transform: pipeTr})

		pipeTr = cam.Translate(pipeX, 32+pipeY+64)
		screen.DrawImage(g.pipe, &orion.DrawImageOptions{Transform: pipeTr})
	}

	// draw foreground tiles
	g.drawTiles(screen, g.foreground, cam, float32(256-g.foreground.Height()), 1.0)

	// draw debug overlay
	orion.DebugOverlay.Draw(screen)
}

func (g *Game) DrawToSurface(surface, offscreen *orion.Image) {
	orion.DefaultDrawToSurface(surface, offscreen, wgpu.FilterModeNearest)
}

func (g *Game) drawTiles(target *orion.Image, tile *orion.Image, cam glm.Mat3[float32], y, parallaxScale float32) {
	parallaxOffset := g.xOffset * (1 - parallaxScale)

	// calculate the first multiple of tile.Width() that is greater than the left corner we need to draw.
	firstTile := uint32((g.xOffset - parallaxOffset) / float32(tile.Width()))

	tileCount := 480/tile.Width() + 2
	for idx := range tileCount {
		xTile := float32((firstTile+idx)*tile.Width()) + parallaxOffset
		target.DrawImage(tile, &orion.DrawImageOptions{
			Transform: cam.Mul(glm.TranslationMat3(xTile, y)),
		})
	}
}

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	err := orion.RunGame(orion.RunGameOptions{
		Game:         &Game{},
		WindowWidth:  480 * 3,
		WindowHeight: 256 * 3,
		WindowTitle:  "Flappy",
	})

	orion.Handle(err, "run game")
}
