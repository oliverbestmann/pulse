package commands

import (
	_ "embed"
	"fmt"
	"log/slog"
	"unsafe"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse"
)

//go:embed sprite2d.wgsl
var spriteShaderCode string

// maximum number of sprite vertices to render in one batchConfig.
const maxSpriteInstances = 128 * 1024

type spriteVertexUniforms struct {
	targetTextureSize glm.Vec2f
	sourceTextureSize glm.Vec2f
}

type spriteBatchConfig struct {
	target              *pulse.Texture
	texture             *wgpu.TextureView
	filterMode          wgpu.FilterMode
	blendState          wgpu.BlendState
	addressModeU        wgpu.AddressMode
	addressModeV        wgpu.AddressMode
	sourceTextureWidth  uint32
	sourceTextureHeight uint32
	shader              string
}

type spriteInstance struct {
	// Color to tint the sprite with
	Color pulse.Color

	// first and second row of the transposed affine
	ModelTransposedCol0 glm.Vec3f
	ModelTransposedCol1 glm.Vec3f

	// Source region within the source texture (x, y, w, h).
	// The source rectangle maps to uv 0 to 1.
	SourceRegion glm.Vec4uh

	// Target region to draw to in target texture (x, y, w, h)
	// The sprites vertex coordinates are interpreted relative to x, y
	// and are clipped to the target region. The sprite is a square from 0 to 1 and
	// transformed with the model matrix first.
	TargetRegion glm.Vec4uh
}

type SpriteCommand struct {
	ctx *pulse.Context

	pipelineCache *pulse.PipelineCache[spritePipelineConfig]

	instances    []spriteInstance
	bufInstances *wgpu.Buffer
	bufIndices   *wgpu.Buffer

	bufVertexUniforms *wgpu.Buffer

	batchConfig spriteBatchConfig
}

func NewSpriteCommand(ctx *pulse.Context) (*SpriteCommand, error) {
	// create a vertex buffer
	bufInstances, err := ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Sprite.Instances",
		Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
		Size:  uint64(unsafe.Sizeof(spriteInstance{})) * maxSpriteInstances,
	})

	if err != nil {
		return nil, fmt.Errorf("create instance buffer: %w", err)
	}

	bufIndices, err := ctx.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label:    "Sprite.Indices",
		Contents: wgpu.ToBytes([]uint16{2, 0, 1, 1, 3, 2}),
		Usage:    wgpu.BufferUsageIndex,
	})

	if err != nil {
		return nil, fmt.Errorf("create index buffer: %w", err)
	}

	bufVertexUniforms, err := ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Sprite.VertexUniforms",
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		Size:  uint64(unsafe.Sizeof(spriteVertexUniforms{})),
	})

	if err != nil {
		return nil, fmt.Errorf("create view transform uniform: %w", err)
	}

	p := &SpriteCommand{
		ctx:               ctx,
		bufInstances:      bufInstances,
		bufIndices:        bufIndices,
		bufVertexUniforms: bufVertexUniforms,
	}

	p.pipelineCache = pulse.NewPipelineCache[spritePipelineConfig](ctx)

	return p, nil
}

type DrawSpriteOptions struct {
	Transform    glm.Mat3f
	Color        pulse.Color
	FilterMode   wgpu.FilterMode
	BlendState   wgpu.BlendState
	AddressModeU wgpu.AddressMode
	AddressModeV wgpu.AddressMode

	// shader code, use default if empty
	Shader string
}

func (p *SpriteCommand) Draw(dest *pulse.Texture, source *pulse.Texture, opts DrawSpriteOptions) error {
	if opts.Shader == "" {
		opts.Shader = spriteShaderCode
	}

	batchConfig := spriteBatchConfig{
		target:              dest.Root(),
		texture:             source.Root().SourceView(),
		sourceTextureWidth:  source.Root().Width(),
		sourceTextureHeight: source.Root().Height(),
		filterMode:          opts.FilterMode,
		blendState:          opts.BlendState,
		addressModeU:        opts.AddressModeU,
		addressModeV:        opts.AddressModeV,
		shader:              opts.Shader,
	}

	requireFlush := p.batchConfig != batchConfig ||
		len(p.instances)+1 > maxSpriteInstances

	if requireFlush {
		if err := p.Flush(); err != nil {
			return fmt.Errorf("flush: %w", err)
		}

		p.batchConfig = batchConfig
	}

	sx, sy := source.Offset().XY()
	sw, sh := source.Size().XY()

	dx, dy := dest.Offset().XY()
	dw, dh := dest.Size().XY()

	p.instances = append(p.instances, spriteInstance{
		Color:               opts.Color,
		ModelTransposedCol0: opts.Transform.Row(0),
		ModelTransposedCol1: opts.Transform.Row(1),

		SourceRegion: glm.Vec4uh{
			uint16(sx),
			uint16(sy),
			uint16(sw),
			uint16(sh),
		},

		TargetRegion: glm.Vec4uh{
			uint16(dx),
			uint16(dy),
			uint16(dw),
			uint16(dh),
		},
	})

	return nil
}

type DrawSpriteFromGPUOptions struct {
	Buffer        *wgpu.Buffer
	InstanceCount uint

	FilterMode   wgpu.FilterMode
	BlendState   wgpu.BlendState
	AddressModeU wgpu.AddressMode
	AddressModeV wgpu.AddressMode
	Shader       string
}

func (p *SpriteCommand) DrawFromGPU(dest *pulse.Texture, source *pulse.Texture, opts DrawSpriteFromGPUOptions) error {
	if err := p.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	if opts.Shader == "" {
		opts.Shader = spriteShaderCode
	}

	p.batchConfig = spriteBatchConfig{
		target:              dest.Root(),
		texture:             source.Root().SourceView(),
		sourceTextureWidth:  source.Root().Width(),
		sourceTextureHeight: source.Root().Height(),
		filterMode:          opts.FilterMode,
		blendState:          opts.BlendState,
		addressModeU:        opts.AddressModeU,
		addressModeV:        opts.AddressModeV,
		shader:              opts.Shader,
	}

	return p.flushWith(opts.Buffer, uint32(opts.InstanceCount), nil)

}

func (p *SpriteCommand) Flush() error {
	if len(p.instances) == 0 {
		return nil
	}

	slog.Debug("Rendering sprites", slog.Int("instanceCount", len(p.instances)))

	err := p.ctx.WriteBuffer(p.bufInstances, 0, wgpu.ToBytes(p.instances))
	if err != nil {
		return fmt.Errorf("update instance buffer: %w", err)
	}

	x0 := p.instances[0].TargetRegion[0]
	y0 := p.instances[0].TargetRegion[1]
	x1 := x0 + p.instances[0].TargetRegion[2]
	y1 := y0 + p.instances[0].TargetRegion[3]

	for idx := range p.instances {
		x0 = min(x0, p.instances[idx].TargetRegion[0])
		y0 = min(y0, p.instances[idx].TargetRegion[1])
		x1 = max(x1, p.instances[idx].TargetRegion[0]+p.instances[idx].TargetRegion[2])
		y1 = max(y1, p.instances[idx].TargetRegion[1]+p.instances[idx].TargetRegion[3])
	}

	rect := pulse.Rectangle2u{
		Min: glm.Vec2[uint32]{
			uint32(x0), uint32(y0),
		},
		Max: glm.Vec2[uint32]{
			uint32(x1), uint32(y1),
		},
	}

	return p.flushWith(p.bufInstances, uint32(len(p.instances)), &rect)
}

func (p *SpriteCommand) flushWith(instances *wgpu.Buffer, instanceCount uint32, scissorRect *pulse.Rectangle2u) error {
	defer p.reset()

	batchConfig := p.batchConfig

	descSampler := wgpu.SamplerDescriptor{
		Label:         "UserTex-Sampler",
		AddressModeU:  batchConfig.addressModeU,
		AddressModeV:  batchConfig.addressModeV,
		AddressModeW:  wgpu.AddressModeUndefined,
		MagFilter:     batchConfig.filterMode,
		MinFilter:     batchConfig.filterMode,
		MipmapFilter:  wgpu.MipmapFilterModeLinear,
		LodMinClamp:   1,
		LodMaxClamp:   1,
		MaxAnisotropy: 1,
	}

	sampler, err := pulse.CachedSampler(p.ctx.Device, descSampler)
	if err != nil {
		return err
	}

	pipelineConfig := spritePipelineConfig{
		TargetFormat:      batchConfig.target.Format(),
		TargetSampleCount: batchConfig.target.SampleCount(),
		BlendState:        batchConfig.blendState,
		ShaderSource:      batchConfig.shader,
	}

	pc, err := p.pipelineCache.Get(pipelineConfig)
	if err != nil {
		return fmt.Errorf("get new pipeline: %w", err)
	}

	bindGroup, err := p.ctx.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: pc.GetBindGroupLayout(0),
		Entries: []wgpu.BindGroupEntry{
			{
				Binding:     0,
				TextureView: batchConfig.texture,
			},
			{
				Binding: 1,
				Sampler: sampler,
			},
			{
				Binding: 2,
				Buffer:  p.bufVertexUniforms,
				Size:    wgpu.WholeSize,
			},
		},
	})

	if err != nil {
		return err
	}

	defer bindGroup.Release()

	// prepare uniforms to upload
	uni := spriteVertexUniforms{
		targetTextureSize: glm.Vec2f{
			float32(batchConfig.target.Width()),
			float32(batchConfig.target.Height()),
		},

		sourceTextureSize: glm.Vec2f{
			float32(batchConfig.sourceTextureWidth),
			float32(batchConfig.sourceTextureHeight),
		},
	}

	// upload to uniform buffer
	err = p.ctx.WriteBuffer(p.bufVertexUniforms, 0, pulse.AsByteSlice(&uni))
	if err != nil {
		return fmt.Errorf("update view transform buffer: %w", err)
	}

	// create command encoder to prepare render pass
	encoder, err := p.ctx.CreateCommandEncoder(nil)
	if err != nil {
		return err
	}
	defer encoder.Release()

	view, resolveTarget := batchConfig.target.Views()

	pass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "RenderPassSprite",
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:          view,
				ResolveTarget: resolveTarget,
				LoadOp:        wgpu.LoadOpLoad,
				StoreOp:       wgpu.StoreOpStore,
			},
		},
	})

	defer func() {
		if pass != nil {
			pass.Release()
		}
	}()

	if scissorRect != nil {
		pass.SetScissorRect(scissorRect.XYWH())
	}

	pass.SetPipeline(pc.Pipeline)
	pass.SetBindGroup(0, bindGroup, nil)
	pass.SetVertexBuffer(0, instances, 0, wgpu.WholeSize)
	pass.SetIndexBuffer(p.bufIndices, wgpu.IndexFormatUint16, 0, wgpu.WholeSize)
	pass.DrawIndexed(6, instanceCount, 0, 0, 0)

	if err := pass.End(); err != nil {
		return err
	}

	// must release pass before finishing the encoder
	pass.Release()
	pass = nil

	cmdBuffer, err := encoder.Finish(nil)
	if err != nil {
		return err
	}

	defer cmdBuffer.Release()

	p.ctx.Submit(cmdBuffer)

	return nil
}

type spritePipelineConfig struct {
	TargetFormat      wgpu.TextureFormat
	BlendState        wgpu.BlendState
	TargetSampleCount uint32
	ShaderSource      string
}

func (conf spritePipelineConfig) Specialize(dev *wgpu.Device) (*wgpu.RenderPipeline, error) {
	slog.Info(
		"Create RenderPipeline for sprites",
		slog.Any("config", conf.TargetFormat),
		slog.Any("sampleCount", conf.TargetSampleCount),
	)

	shader, err := dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:      "Sprite2D.ShaderSource",
		WGSLSource: &wgpu.ShaderSourceWGSL{Code: conf.ShaderSource},
	})
	if err != nil {
		return nil, fmt.Errorf("compile sprite shader: %w", err)
	}

	defer shader.Release()

	desc := &wgpu.RenderPipelineDescriptor{
		Label: fmt.Sprintf("Sprite2D.%s", conf.TargetFormat),
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{
				{
					StepMode:    wgpu.VertexStepModeInstance,
					ArrayStride: uint64(unsafe.Sizeof(spriteInstance{})),
					Attributes: []wgpu.VertexAttribute{
						{
							// color
							Format:         wgpu.VertexFormatFloat32x4,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.Color)),
							ShaderLocation: 0,
						},
						{
							// transform, row0
							Format:         wgpu.VertexFormatFloat32x3,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.ModelTransposedCol0)),
							ShaderLocation: 1,
						},
						{
							// transform, row1
							Format:         wgpu.VertexFormatFloat32x3,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.ModelTransposedCol1)),
							ShaderLocation: 2,
						},
						{
							// uv pos
							Format:         wgpu.VertexFormatUint32x2,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.SourceRegion)),
							ShaderLocation: 3,
						},
						{
							// uv scale
							Format:         wgpu.VertexFormatUint32x2,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.TargetRegion)),
							ShaderLocation: 4,
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

	pipeline, err := dev.CreateRenderPipeline(desc)
	if err != nil {
		return nil, fmt.Errorf("build sprite pipeline: %w", err)
	}

	return pipeline, nil
}

func (p *SpriteCommand) reset() {
	p.instances = p.instances[:0]
	p.batchConfig = spriteBatchConfig{}
}
