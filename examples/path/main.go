package main

import (
	"math"
	"time"

	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/orion"
	"github.com/oliverbestmann/pulse/orion/vector"
	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/webgpu/wgpu"
)

type Game struct {
}

func (g Game) Layout(surfaceWidth, surfaceHeight uint32) orion.LayoutOptions {
	return orion.LayoutOptions{
		Width:  surfaceWidth,
		Height: surfaceHeight,
		MSAA:   true,
	}
}

func (g Game) Initialize() error {
	orion.DebugOverlay.Enable(true)
	return nil
}

func (g Game) Update() error {
	return nil
}

func (g Game) Draw(screen *orion.Image) {
	scale := float32(screen.Width()) / 768.0
	toScreen := glm.Mat3f{}.Scale(scale, scale)

	screen.Clear(pulse.ColorBlack)

	var path vector.Path
	path.MoveTo(glm.Vec2f{200, 200})
	path.LineTo(glm.Vec2f{300, 200 + jitter(40)})
	path.QuadCurveTo(glm.Vec2f{100 + jitter(50), 300}, glm.Vec2f{250, 400})
	path.CubicCurveTo(glm.Vec2f{100, 400 + jitter(20)}, glm.Vec2f{100 + jitter(30), 200}, glm.Vec2f{200, 200})

	vector.FillPath(screen, path, &vector.FillPathOptions{
		Transform:  toScreen.Translate(0, -200).Scale(2.0, 2.0),
		ColorScale: pulse.ColorLinearRGBA(1.0, 0.3, 0.6, 1.0),
	})

	smallScreen := screen.SubImage(200, 200, screen.Width()-400, screen.Height()-400)

	vector.StrokePath(smallScreen, path, &vector.StrokePathOptions{
		Transform:  glm.Mat3f{}.Translate(-200, -200).Mul(toScreen).Translate(0, -200).Scale(2.0, 2.0),
		ColorScale: pulse.ColorLinearRGBA(1, 1, 1, 1.0),
		Thickness:  1,
	})

	// draw a linear gradient
	points := []orion.Vertex2d{
		{
			Position: glm.Vec2f{},
			Color:    pulse.ColorLinearRGBA(0, 0, 0, 1),
		},
		{
			Position: glm.Vec2f{0, 200},
			Color:    pulse.ColorLinearRGBA(0, 0, 0, 1),
		},
		{
			Position: glm.Vec2f{800, 200},
			Color:    pulse.ColorLinearRGBA(1, 1, 1, 1),
		},
	}

	screen.DrawTriangles(points, &orion.DrawTrianglesOptions{})

	orion.DebugOverlay.Draw(screen)
}

func (g Game) DrawToSurface(surface, offscreen *orion.Image) {
	orion.DefaultDrawToSurface(surface, offscreen, wgpu.FilterModeLinear)
}

func main() {
	err := orion.RunGame(orion.RunGameOptions{
		Game:            &Game{},
		WindowWidth:     1024,
		WindowHeight:    768,
		WindowTitle:     "Paths",
		WindowResizable: true,
	})

	if err != nil {
		panic(err)
	}
}

func jitter(i float32) float32 {
	t := float64(time.Now().UnixMilli()) / 1000.0
	return float32(math.Sin(t)) * i
}
