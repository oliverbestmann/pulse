package pulse

import (
	_ "embed"
	"fmt"
	"log/slog"
	"unsafe"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
)

//go:embed sprite2d.wgsl
var spriteShaderCode string

// maximum number of sprite vertices to render in one batchConfig.
const maxSpriteInstances = 128 * 1024

type spriteBatchConfig struct {
	target       *RenderTarget
	texture      *wgpu.TextureView
	filterMode   wgpu.FilterMode
	blendState   wgpu.BlendState
	addressModeU wgpu.AddressMode
	addressModeV wgpu.AddressMode
	sourceWidth  uint32
	sourceHeight uint32
	shader       string
}

type spriteInstance struct {
	Color Color

	UVOffset glm.Vec2f
	UVScale  glm.Vec2f

	// first and second row of the transposed affine
	ModelTransposedCol0 glm.Vec3f
	ModelTransposedCol1 glm.Vec3f
}

type SpriteCommand struct {
	ctx *Context

	pipelineCache *PipelineCache[spritePipelineConfig]

	instances    []spriteInstance
	bufInstances *wgpu.Buffer
	bufIndices   *wgpu.Buffer

	bufViewTransform  *wgpu.Buffer
	bufLocalTransform *wgpu.Buffer

	batchConfig spriteBatchConfig
}

func NewSpriteCommand(ctx *Context) (*SpriteCommand, error) {
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

	bufViewTransform, err := ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Sprite.ViewUniform",
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		Size:  uint64(unsafe.Sizeof([4 * 4]float32{})),
	})

	if err != nil {
		return nil, fmt.Errorf("create view transform uniform: %w", err)
	}

	bufLocalTransform, err := ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Sprite.LocalUniform",
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		Size:  uint64(unsafe.Sizeof([4 * 4]float32{})),
	})

	if err != nil {
		return nil, fmt.Errorf("create local transform uniform: %w", err)
	}

	p := &SpriteCommand{
		ctx:               ctx,
		bufInstances:      bufInstances,
		bufIndices:        bufIndices,
		bufViewTransform:  bufViewTransform,
		bufLocalTransform: bufLocalTransform,
	}

	p.pipelineCache = NewPipelineCache[spritePipelineConfig](ctx)

	return p, nil
}

type DrawSpriteOptions struct {
	Transform    glm.Mat3f
	Color        Color
	FilterMode   wgpu.FilterMode
	BlendState   wgpu.BlendState
	AddressModeU wgpu.AddressMode
	AddressModeV wgpu.AddressMode

	// shader code, use default if empty
	Shader string
}

func (p *SpriteCommand) Draw(dest *RenderTarget, source *Texture, opts DrawSpriteOptions) error {
	if opts.Shader == "" {
		opts.Shader = spriteShaderCode
	}

	batchConfig := spriteBatchConfig{
		target:       dest,
		texture:      source.SourceView(),
		sourceWidth:  source.Width(),
		sourceHeight: source.Height(),
		filterMode:   opts.FilterMode,
		blendState:   opts.BlendState,
		addressModeU: opts.AddressModeU,
		addressModeV: opts.AddressModeV,
		shader:       opts.Shader,
	}

	requireFlush := p.batchConfig != batchConfig ||
		len(p.instances)+1 > maxSpriteInstances

	if requireFlush {
		if err := p.Flush(); err != nil {
			return fmt.Errorf("flush: %w", err)
		}

		p.batchConfig = batchConfig
	}

	p.instances = append(p.instances, spriteInstance{
		Color:               opts.Color,
		UVOffset:            source.UVOffset(),
		UVScale:             source.UVScale(),
		ModelTransposedCol0: opts.Transform.Row(0),
		ModelTransposedCol1: opts.Transform.Row(1),
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

func (p *SpriteCommand) DrawFromGPU(dest *RenderTarget, source *Texture, opts DrawSpriteFromGPUOptions) error {
	if err := p.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	if opts.Shader == "" {
		opts.Shader = spriteShaderCode
	}

	p.batchConfig = spriteBatchConfig{
		target:       dest,
		texture:      source.SourceView(),
		sourceWidth:  source.Width(),
		sourceHeight: source.Height(),
		filterMode:   opts.FilterMode,
		blendState:   opts.BlendState,
		addressModeU: opts.AddressModeU,
		addressModeV: opts.AddressModeV,
		shader:       opts.Shader,
	}

	return p.flushWith(opts.Buffer, uint32(opts.InstanceCount))

}

func (p *SpriteCommand) Flush() error {
	if len(p.instances) == 0 {
		return nil
	}

	slog.Debug("Rendering sprites", slog.Int("instanceCount", len(p.instances)))

	queue := p.ctx.GetQueue()
	defer queue.Release()

	err := queue.WriteBuffer(p.bufInstances, 0, wgpu.ToBytes(p.instances))
	if err != nil {
		return fmt.Errorf("update instance buffer: %w", err)
	}

	return p.flushWith(p.bufInstances, uint32(len(p.instances)))
}

func (p *SpriteCommand) flushWith(instances *wgpu.Buffer, instanceCount uint32) error {
	defer p.reset()

	batchConfig := p.batchConfig

	queue := p.ctx.GetQueue()
	defer queue.Release()

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

	sampler, err := p.ctx.CreateSampler(&descSampler)
	if err != nil {
		return err
	}

	defer sampler.Release()

	pipelineConfig := spritePipelineConfig{
		TargetFormat:      batchConfig.target.Format,
		TargetSampleCount: batchConfig.target.SampleCount,
		BlendState:        batchConfig.blendState,
		ShaderSource:      batchConfig.shader,
	}

	pipeline, err := p.pipelineCache.Get(pipelineConfig)
	if err != nil {
		return fmt.Errorf("get new pipeline: %w", err)
	}

	bindGroupLayout := pipeline.GetBindGroupLayout(0)
	defer bindGroupLayout.Release()

	bindGroup, err := p.ctx.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Layout: bindGroupLayout,
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
				Buffer:  p.bufViewTransform,
				Size:    wgpu.WholeSize,
			},
			{
				Binding: 3,
				Buffer:  p.bufLocalTransform,
				Size:    wgpu.WholeSize,
			},
		},
	})

	if err != nil {
		return err
	}

	defer bindGroup.Release()

	// build a new view transform
	vw, vh := batchConfig.target.Width, batchConfig.target.Height
	viewTransform := glm.ScaleMat3(1/float32(vw), 1/float32(vh))

	viewTransformValues := viewTransform.ToWGPU()
	err = queue.WriteBuffer(p.bufViewTransform, 0, AsByteSlice(&viewTransformValues))
	if err != nil {
		return fmt.Errorf("update view transform buffer: %w", err)
	}

	// scale by size of the source image
	sw, sh := float32(batchConfig.sourceWidth), float32(batchConfig.sourceHeight)
	localTransform := glm.ScaleMat3(sw, sh)

	localTransformValues := localTransform.ToWGPU()
	err = queue.WriteBuffer(p.bufLocalTransform, 0, AsByteSlice(&localTransformValues))
	if err != nil {
		return fmt.Errorf("update local transform buffer: %w", err)
	}

	encoder, err := p.ctx.CreateCommandEncoder(nil)
	if err != nil {
		return err
	}
	defer encoder.Release()

	pass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "RenderPassSprite",
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:          batchConfig.target.View,
				ResolveTarget: batchConfig.target.ResolveTarget,
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

	pass.SetPipeline(pipeline)
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

	queue.Submit(cmdBuffer)

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
					ArrayStride: uint64(unsafe.Sizeof(spriteInstance{})),
					StepMode:    wgpu.VertexStepModeInstance,
					Attributes: []wgpu.VertexAttribute{
						{
							// color
							Format:         wgpu.VertexFormatFloat32x4,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.Color)),
							ShaderLocation: 0,
						},
						{
							// uv pos
							Format:         wgpu.VertexFormatFloat32x2,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.UVOffset)),
							ShaderLocation: 1,
						},
						{
							// uv scale
							Format:         wgpu.VertexFormatFloat32x2,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.UVScale)),
							ShaderLocation: 2,
						},
						{
							// transform, row0
							Format:         wgpu.VertexFormatFloat32x3,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.ModelTransposedCol0)),
							ShaderLocation: 3,
						},
						{
							// transform, row1
							Format:         wgpu.VertexFormatFloat32x3,
							Offset:         uint64(unsafe.Offsetof(spriteInstance{}.ModelTransposedCol1)),
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
