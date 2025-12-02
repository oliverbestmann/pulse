package vector

import (
	"unsafe"

	"github.com/oliverbestmann/earcut-go"
	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/orion"
	"github.com/oliverbestmann/webgpu/wgpu"
)

type FillPathOptions struct {
	Transform  glm.Mat3f
	ColorScale orion.ColorScale
	BlendState wgpu.BlendState
	Shader     string
}

func FillPath(target *orion.Image, path Path, opts *FillPathOptions) {
	if opts == nil {
		opts = &FillPathOptions{}
	}

	unitScale := calculateUnitScale(opts.Transform)

	points := earcutPointsOf(path.Contour(unitScale))
	points, indices := earcut.Triangulate(points, nil)

	// convert to vertices
	vertices := make([]orion.Vertex2d, len(indices))
	for i := 0; i < len(indices); i += 3 {
		vertices[i] = orion.Vertex2d{
			Position:   toVec(points[indices[i]]),
			ColorScale: opts.ColorScale,
		}

		vertices[i+1] = orion.Vertex2d{
			Position:   toVec(points[indices[i+1]]),
			ColorScale: opts.ColorScale,
		}

		vertices[i+2] = orion.Vertex2d{
			Position:   toVec(points[indices[i+2]]),
			ColorScale: opts.ColorScale,
		}
	}

	target.DrawTriangles(vertices, &orion.DrawTrianglesOptions{
		Transform:  opts.Transform,
		ColorScale: opts.ColorScale,
		BlendState: opts.BlendState,
		Shader:     opts.Shader,
	})
}

func calculateUnitScale(transform glm.Mat3f) float32 {
	a := transform.Transform2(glm.Vec2f{1, 0})
	b := transform.Transform2(glm.Vec2f{0, 0})
	return max(0.1, 0.5/a.Sub(b).Length())
}

type StrokePathOptions struct {
	Transform  glm.Mat3f
	ColorScale orion.ColorScale
	BlendState wgpu.BlendState
	Thickness  float32
}

var drawLines *drawLinesCommand

func StrokePath(target *orion.Image, path Path, opts *StrokePathOptions) {
	if opts == nil {
		opts = &StrokePathOptions{}
	}

	unitScale := calculateUnitScale(opts.Transform)
	points := path.Contour(unitScale)

	if drawLines == nil {
		drawLines = &drawLinesCommand{}
		drawLines.Init()
	}

	orion.SwitchToCommand(drawLines)

	err := drawLines.Draw(target.Texture(), points, *opts)
	orion.Handle(err, "stroke path")
}

func toVec(point earcut.Point[float32]) glm.Vec2f {
	return glm.Vec2f{point.X, point.Y}
}

func earcutPointsOf(vecs []glm.Vec2f) []earcut.Point[float32] {
	// convert to earcut points. the layout of both types are the
	// same, so we can just cast the slice
	ptr := (*earcut.Point[float32])(unsafe.Pointer(unsafe.SliceData(vecs)))
	return unsafe.Slice(ptr, len(vecs))
}
