package main

import (
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/orion"
	"github.com/oliverbestmann/go3d/orion/vector"
	"github.com/oliverbestmann/webgpu/wgpu"
)

type Game struct {
}

func (g Game) Layout(surfaceWidth, surfaceHeight uint32) orion.LayoutOptions {
	return orion.LayoutOptions{
		Width:  1024,
		Height: 768,
		MSAA:   true,
	}
}

func (g Game) Initialize() error {
	return nil
}

func (g Game) Update() error {
	return nil
}

func (g Game) Draw(screen *orion.Image) {
	var path vector.Path
	path.MoveTo(glm.Vec2f{200, 200})
	path.LineTo(glm.Vec2f{300, 200})
	path.QuadCurveTo(glm.Vec2f{100, 300}, glm.Vec2f{250, 400})
	path.CubicCurveTo(glm.Vec2f{100, 400}, glm.Vec2f{100, 200}, glm.Vec2f{200, 200})

	vector.FillPath(screen, path, &vector.FillPathOptions{
		Transform:  glm.Mat3f{}.Translate(0, -200).Scale(2.0, 2.0),
		ColorScale: orion.ColorScaleRGBA(1.0, 0.3, 0.6, 1.0),
	})
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
