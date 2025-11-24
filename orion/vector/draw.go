package vector

import (
	"unsafe"

	"github.com/oliverbestmann/earcut-go"
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

func toVec(point earcut.Point[float32]) glm.Vec2f {
	return glm.Vec2f{point.X, point.Y}
}

func earcutPointsOf(vecs []glm.Vec2f) []earcut.Point[float32] {
	// convert to earcut points. the layout of both types are the
	// same, so we can just cast the slice
	ptr := (*earcut.Point[float32])(unsafe.Pointer(unsafe.SliceData(vecs)))
	return unsafe.Slice(ptr, len(vecs))
}
