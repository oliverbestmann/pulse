package vector

import (
	_ "embed"
	"fmt"
	"unsafe"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/orion"
	"github.com/oliverbestmann/go3d/pulse"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed lines.wgsl
var lineShader string

const circleTriangleCount = 32

type drawLinesCommand struct {
	cache *pulse.PipelineCache[pipelineStub]

	pointsBuf  *wgpu.Buffer
	configsBuf *wgpu.Buffer
}

func (d *drawLinesCommand) Flush() error {
	return nil
}

func (d *drawLinesCommand) Init() {
	ctx := orion.CurrentContext()

	d.cache = pulse.NewPipelineCache[pipelineStub](ctx)

	d.configsBuf = orion.CreateBuffer(wgpu.BufferDescriptor{
		Label: "LineConfig",
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		Size:  128,
	})

	d.pointsBuf = orion.CreateBuffer(wgpu.BufferDescriptor{
		Label: "LinePoints",
		Usage: wgpu.BufferUsageStorage | wgpu.BufferUsageCopyDst,
		Size:  1024 * 1024,
	})
}

func (d *drawLinesCommand) Draw(target *pulse.Texture, points []glm.Vec2f, opts StrokePathOptions) error {
	const maxPointsPerDrawCall = int(1024 * 1024 / unsafe.Sizeof(glm.Vec2f{}))

	if len(points) > maxPointsPerDrawCall {
		// recurse with rest of points
		if err := d.Draw(target, points[maxPointsPerDrawCall:], opts); err != nil {
			return err
		}

		// draw the first batch of points directly
		points = points[:maxPointsPerDrawCall]
	}

	var blendState = orion.BlendStateDefault
	if opts.BlendState != (wgpu.BlendState{}) {
		blendState = opts.BlendState
	}

	pipelineConf := pipelineStub{
		Target:      target.Root(),
		Blend:       blendState,
		Format:      target.Format(),
		SampleCount: target.SampleCount(),
	}

	toClipSpace := glm.Mat3f{}.
		Translate(-1.0, 1.0).
		Scale(2.0/float32(target.Root().Width()), -2.0/float32(target.Root().Height())).
		Translate(target.Offset().ToVec2f().XY())

	projection := toClipSpace.Mul(opts.Transform)

	// record the line instance
	config := lineConfig{
		Projection:  projection.ToWGPU(),
		Color:       opts.ColorScale.ToColor(),
		Thickness:   opts.Thickness,
		PointsCount: uint32(len(points)),
	}

	pipeline, err := d.cache.Get(pipelineConf)
	if err != nil {
		return fmt.Errorf("get pipeline: %w", err)
	}

	dev := orion.CurrentContext()

	stencilView := d.getStencilTex(target.Root())

	enc, err := dev.CreateCommandEncoder(nil)
	if err != nil {
		return fmt.Errorf("create command encoder: %w", err)
	}

	view, resolveTarget := target.RenderViews()

	pass := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:          view,
				ResolveTarget: resolveTarget,
				LoadOp:        wgpu.LoadOpLoad,
				StoreOp:       wgpu.StoreOpStore,
			},
		},
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:           stencilView,
			StencilLoadOp:  wgpu.LoadOpClear,
			StencilStoreOp: wgpu.StoreOpStore,
		},
	})

	bindGroup, err := dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: pipeline.GetBindGroupLayout(0),
		Entries: []wgpu.BindGroupEntry{
			{
				Binding: 0,
				Buffer:  d.configsBuf,
				Size:    wgpu.WholeSize,
			},
			{
				Binding: 1,
				Buffer:  d.pointsBuf,
				Size:    wgpu.WholeSize,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("create bindGroup: %w", err)
	}

	defer bindGroup.Release()

	x, y := target.Offset().XY()
	w, h := target.Size().XY()

	pass.SetPipeline(pipeline.Pipeline)
	pass.SetBindGroup(0, bindGroup, nil)
	pass.SetStencilReference(1)
	pass.SetScissorRect(x, y, w, h)
	pass.Draw(6+circleTriangleCount*3, uint32(len(points)), 0, 0)

	if err := pass.End(); err != nil {
		return fmt.Errorf("end render pass: %w", err)
	}

	pass.Release()

	buf, err := enc.Finish(nil)
	if err != nil {
		return fmt.Errorf("finish command buffer: %w", err)
	}

	defer buf.Release()

	// upload data to gpu and draw
	err = dev.Queue.WriteBuffer(d.configsBuf, 0, wgpu.ToBytes([]lineConfig{config}))
	orion.Handle(err, "upload line config")

	err = dev.Queue.WriteBuffer(d.pointsBuf, 0, wgpu.ToBytes(points))
	orion.Handle(err, "upload line points")

	dev.Queue.Submit(buf)

	return nil
}

func (d *drawLinesCommand) getStencilTex(target *pulse.Texture) *wgpu.TextureView {
	desc := wgpu.TextureDescriptor{
		Label:     "LinesStencil",
		Usage:     wgpu.TextureUsageRenderAttachment,
		Dimension: wgpu.TextureDimension2D,
		Size: wgpu.Extent3D{
			Width:              target.Width(),
			Height:             target.Height(),
			DepthOrArrayLayers: 1,
		},
		Format:        wgpu.TextureFormatStencil8,
		SampleCount:   target.SampleCount(),
		MipLevelCount: 1,
	}

	if cached, ok := stencilTexCache.Get(desc); ok {
		return cached
	}

	dev := orion.CurrentContext()

	texture, err := dev.CreateTexture(&desc)
	orion.Handle(err, "create line stencil texture")

	view, err := texture.CreateView(nil)
	orion.Handle(err, "create line stencil texture view")

	stencilTexCache.Add(desc, view)

	return view
}

type lineConfig struct {
	Projection  [12]float32
	Color       glm.Vec4f
	Thickness   float32
	PointsCount uint32
}

type pipelineStub struct {
	Target *pulse.Texture

	Blend       wgpu.BlendState
	Format      wgpu.TextureFormat
	SampleCount uint32
}

func (d pipelineStub) Specialize(dev *wgpu.Device) (*wgpu.RenderPipeline, error) {
	shader := orion.CreateShaderModule(wgpu.ShaderModuleDescriptor{
		Label:      "LinesShader",
		WGSLSource: &wgpu.ShaderSourceWGSL{Code: lineShader},
	})

	return dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label: "LinesPipeline",
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vertex",
		},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: "fragment",
			Targets: []wgpu.ColorTargetState{
				{
					Format:    d.Format,
					Blend:     &d.Blend,
					WriteMask: wgpu.ColorWriteMaskAll,
				},
			},
		},
		DepthStencil: &wgpu.DepthStencilState{
			Format: wgpu.TextureFormatStencil8,
			StencilFront: wgpu.StencilFaceState{
				// draw only if reference is greater than stencil value.
				//  set stencil value to reference if drawn
				Compare: wgpu.CompareFunctionGreater,
				PassOp:  wgpu.StencilOperationReplace,
				FailOp:  wgpu.StencilOperationKeep,
			},

			StencilBack: wgpu.StencilFaceState{
				Compare: wgpu.CompareFunctionGreater,
				PassOp:  wgpu.StencilOperationReplace,
				FailOp:  wgpu.StencilOperationKeep,
			},

			StencilWriteMask: 0xff,
			StencilReadMask:  0xff,
		},
		Primitive: wgpu.PrimitiveState{
			Topology: wgpu.PrimitiveTopologyTriangleList,
			CullMode: wgpu.CullModeNone,
		},
		Multisample: wgpu.MultisampleState{
			Count:                  d.SampleCount,
			Mask:                   0xffffffff,
			AlphaToCoverageEnabled: false,
		},
	})
}

var stencilTexCache, _ = lru.NewWithEvict[wgpu.TextureDescriptor, *wgpu.TextureView](4, evictStencilTex)

func evictStencilTex(_ wgpu.TextureDescriptor, view *wgpu.TextureView) {
	view.Release()
}
