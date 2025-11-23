package vector

import (
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/orion"
	"github.com/oliverbestmann/webgpu/wgpu"
)

type FillPathOptions struct {
	Transform  glm.Mat3f
	ColorScale orion.ColorScale
	BlendState wgpu.BlendState
	Shader     string
}

func FillPath(target *orion.Image, path Path, opts *FillPathOptions) {
	a := opts.Transform.Transform2(glm.Vec2f{1, 0})
	b := opts.Transform.Transform2(glm.Vec2f{0, 0})
	unitScale := max(0.1, 1.0/a.Sub(b).Length())

	points := path.Contour(unitScale)

	var vertices []orion.Vertex2d

	for idx := 1; idx < len(points); idx++ {
		a := orion.Vertex2d{
			Position:   points[0],
			ColorScale: opts.ColorScale,
		}

		b := orion.Vertex2d{
			Position:   points[idx],
			ColorScale: opts.ColorScale,
		}

		c := orion.Vertex2d{
			Position:   points[idx-1],
			ColorScale: opts.ColorScale,
		}

		vertices = append(vertices, a, b, c)
	}

	target.DrawTriangles(vertices, &orion.DrawTrianglesOptions{
		Transform:  opts.Transform,
		ColorScale: opts.ColorScale,
		BlendState: opts.BlendState,
		Shader:     opts.Shader,
	})
}
