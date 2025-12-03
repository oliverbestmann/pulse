package commands

import (
	_ "embed"
	"fmt"
	"log/slog"
	"structs"
	"unsafe"

	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed mesh2d.wgsl
var mesh2dShaderCode string

// maximum number of vertices to render in one batch
const maxMeshVertices = 128 * 1024 * 3

type mesh2dBatchConfig struct {
	target     *pulse.Texture
	blendState wgpu.BlendState
	shader     string
}

type MeshVertex struct {
	_ structs.HostLayout

	Position glm.Vec2f
	Color    pulse.Color
}

type Mesh2dCommand struct {
	ctx *pulse.Context

	pipelineCache *pulse.PipelineCache[mesh2dRenderPipeline]

	vertices    []MeshVertex
	buvVertices *wgpu.Buffer

	batchConfig mesh2dBatchConfig
}

func NewMesh2dCommand(ctx *pulse.Context) *Mesh2dCommand {
	// create a vertex buffer
	buvVertices := ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Mesh2d.Vertices",
		Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
		Size:  uint64(unsafe.Sizeof(MeshVertex{})) * maxMeshVertices,
	})

	p := &Mesh2dCommand{
		ctx:         ctx,
		buvVertices: buvVertices,
	}

	p.pipelineCache = pulse.NewPipelineCache[mesh2dRenderPipeline](ctx)

	return p
}

type DrawMesh2dOptions struct {
	Transform  glm.Mat3f
	BlendState wgpu.BlendState
	Color      glm.Vec4f
	Vertices   []MeshVertex
	// shader code, use default if empty
	Shader string
}

func (p *Mesh2dCommand) DrawTriangles(target *pulse.Texture, opts DrawMesh2dOptions) {
	if opts.Shader == "" {
		opts.Shader = mesh2dShaderCode
	}

	batchConfig := mesh2dBatchConfig{
		target:     target,
		blendState: opts.BlendState,
		shader:     opts.Shader,
	}

	// build a new view transform
	vw, vh := batchConfig.target.Width(), batchConfig.target.Height()
	viewTransform := glm.ScaleMat3(1/float32(vw), 1/float32(vh))

	// model view matrix
	modelViewTransform := viewTransform.Mul(opts.Transform)

	for idx := 0; idx < len(opts.Vertices); idx += 3 {
		requireFlush := p.batchConfig != batchConfig ||
			len(p.vertices)+3 > maxMeshVertices

		if requireFlush {
			p.Flush()
			p.batchConfig = batchConfig
		}

		for _, v := range opts.Vertices[idx : idx+3] {
			p.vertices = append(p.vertices, MeshVertex{
				Position: modelViewTransform.Transform(v.Position.Extend(1)).Truncate(),
				Color:    v.Color.Mul(opts.Color),
			})
		}
	}
}

func (p *Mesh2dCommand) Flush() {
	defer p.reset()

	if len(p.vertices) == 0 {
		return
	}

	batchConfig := p.batchConfig

	slog.Debug("Rendering triangles", slog.Int("vertexCount", len(p.vertices)))

	p.ctx.WriteBuffer(p.buvVertices, 0, wgpu.ToBytes(p.vertices))

	pipelineConfig := mesh2dRenderPipeline{
		TargetFormat:      batchConfig.target.Format(),
		TargetSampleCount: batchConfig.target.SampleCount(),
		BlendState:        batchConfig.blendState,
		ShaderSource:      batchConfig.shader,
	}

	pc := p.pipelineCache.Get(pipelineConfig)

	encoder := p.ctx.CreateCommandEncoder(nil)
	defer encoder.Release()

	view, resolveTarget := batchConfig.target.RenderViews()

	pass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "RenderPassMesh",
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:          view,
				ResolveTarget: resolveTarget,
				LoadOp:        wgpu.LoadOpLoad,
				StoreOp:       wgpu.StoreOpStore,
			},
		},
	})

	// set target region as clip rect
	sx, sy := batchConfig.target.Offset().XY()
	sw, sh := batchConfig.target.Size().XY()

	pass.SetPipeline(pc.Pipeline)
	pass.SetScissorRect(sx, sy, sw, sh)
	pass.SetVertexBuffer(0, p.buvVertices, 0, wgpu.WholeSize)
	pass.Draw(uint32(len(p.vertices)), 1, 0, 0)
	pass.End()

	cmdBuffer := encoder.Finish(nil)
	defer cmdBuffer.Release()

	p.ctx.Submit(cmdBuffer)
}

type mesh2dRenderPipeline struct {
	TargetFormat      wgpu.TextureFormat
	BlendState        wgpu.BlendState
	TargetSampleCount uint32
	ShaderSource      string
}

func (conf mesh2dRenderPipeline) Specialize(dev *wgpu.Device) *wgpu.RenderPipeline {
	slog.Info(
		"Create RenderPipeline for mesh2d",
		slog.Any("config", conf.TargetFormat),
		slog.Any("sampleCount", conf.TargetSampleCount),
	)

	shader := dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:      "Mesh2D.ShaderSource",
		WGSLSource: &wgpu.ShaderSourceWGSL{Code: conf.ShaderSource},
	})

	defer shader.Release()

	desc := &wgpu.RenderPipelineDescriptor{
		Label: fmt.Sprintf("Mesh2D.%s", conf.TargetFormat),
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{
				{
					ArrayStride: uint64(unsafe.Sizeof(MeshVertex{})),
					StepMode:    wgpu.VertexStepModeVertex,
					Attributes: []wgpu.VertexAttribute{
						{
							// position
							Format:         wgpu.VertexFormatFloat32x2,
							Offset:         uint64(unsafe.Offsetof(MeshVertex{}.Position)),
							ShaderLocation: 0,
						},
						{
							// color
							Format:         wgpu.VertexFormatFloat32x4,
							Offset:         uint64(unsafe.Offsetof(MeshVertex{}.Color)),
							ShaderLocation: 1,
						},
					},
				},
			},
		},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: "fs_main",
			Targets: []wgpu.ColorTargetState{
				{
					Format:    conf.TargetFormat,
					Blend:     &conf.BlendState,
					WriteMask: wgpu.ColorWriteMaskAll,
				},
			},
		},
		Primitive: wgpu.PrimitiveState{
			Topology:  wgpu.PrimitiveTopologyTriangleList,
			FrontFace: wgpu.FrontFaceCCW,
			CullMode:  wgpu.CullModeNone,
		},
		DepthStencil: nil,
		Multisample: wgpu.MultisampleState{
			Count:                  conf.TargetSampleCount,
			Mask:                   0xFFFFFFFF,
			AlphaToCoverageEnabled: false,
		},
	}

	return dev.CreateRenderPipeline(desc)
}

func (p *Mesh2dCommand) reset() {
	p.vertices = p.vertices[:0]
	p.batchConfig = mesh2dBatchConfig{}
}
