package main

import (
	_ "embed"
	"math/rand/v2"

	"github.com/oliverbestmann/pulse/orion"
	"github.com/oliverbestmann/webgpu/wgpu"
)

const Width = 1024 * 2
const Height = 1024 * 2

//go:embed brot.wgsl
var displayShaderSource string

//go:embed brot-compute.wgsl
var computeShaderSource string

type Game struct {
	bufAcc           *wgpu.Buffer
	pipeDisplay      *wgpu.RenderPipeline
	bindGroupDisplay *wgpu.BindGroup
	pipeUpdate       *wgpu.ComputePipeline
	bindGroupUpdate  *wgpu.BindGroup
	bufSeed          *wgpu.Buffer
}

func (g *Game) Layout(surfaceWidth, surfaceHeight uint32) orion.LayoutOptions {
	return orion.LayoutOptions{
		Width:  Width,
		Height: Height,
	}
}

func (g *Game) Initialize() error {
	ctx := orion.CurrentContext()

	g.bufAcc = ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Accumulator",
		Usage: wgpu.BufferUsageStorage,
		Size:  Width * Height * 4,
	})

	g.bufSeed = ctx.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Seed",
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		Size:  16,
	})

	// compile the compute shader
	computeShader := ctx.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:      "ComputeShader",
		WGSLSource: &wgpu.ShaderSourceWGSL{Code: computeShaderSource},
	})

	g.pipeUpdate = ctx.CreateComputePipeline(&wgpu.ComputePipelineDescriptor{
		Label: "Update",
		Compute: wgpu.ProgrammableStageDescriptor{
			Module:     computeShader,
			EntryPoint: "compute",
		},
	})

	g.bindGroupUpdate = ctx.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "UpdateBindGroup",
		Layout: g.pipeUpdate.GetBindGroupLayout(0),
		Entries: []wgpu.BindGroupEntry{
			{
				Binding: 0,
				Buffer:  g.bufAcc,
				Size:    wgpu.WholeSize,
			},
			{
				Binding: 1,
				Buffer:  g.bufSeed,
				Size:    wgpu.WholeSize,
			},
		},
	})

	// compile the display shader
	displayShader := ctx.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:      "DisplayShader",
		WGSLSource: &wgpu.ShaderSourceWGSL{Code: displayShaderSource},
	})

	// create a pipeline for the shader
	g.pipeDisplay = ctx.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label: "Render",
		Vertex: wgpu.VertexState{
			Module:     displayShader,
			EntryPoint: "vertex",
		},
		Fragment: &wgpu.FragmentState{
			Module:     displayShader,
			EntryPoint: "fragment",
			Targets: []wgpu.ColorTargetState{
				{
					Format:    wgpu.TextureFormatRGBA8Unorm,
					Blend:     &wgpu.BlendStateReplace,
					WriteMask: wgpu.ColorWriteMaskAll,
				},
			},
		},
		Multisample: wgpu.MultisampleState{
			Count: 1,
			Mask:  0xffffffff,
		},
	})

	// prepare the bindGroup for display
	g.bindGroupDisplay = ctx.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "RenderBindGroup",
		Layout: g.pipeDisplay.GetBindGroupLayout(0),
		Entries: []wgpu.BindGroupEntry{
			{
				Buffer:  g.bufAcc,
				Size:    wgpu.WholeSize,
				Binding: 0,
			},
		},
	})

	orion.DebugOverlay.Enable(true)

	return nil
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *orion.Image) {
	ctx := orion.CurrentContext()

	view, _ := screen.Texture().RenderViews()

	enc := ctx.CreateCommandEncoder(nil)
	defer enc.Release()

	// encode the compute pass
	{
		pass := enc.BeginComputePass(&wgpu.ComputePassDescriptor{Label: "UpdatePass"})
		pass.SetPipeline(g.pipeUpdate)
		pass.SetBindGroup(0, g.bindGroupUpdate, nil)
		pass.DispatchWorkgroups(256*16, 1, 1)
		pass.End()
	}

	// encode the render pass
	{
		pass := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
			Label: "RenderPass",
			ColorAttachments: []wgpu.RenderPassColorAttachment{
				{
					View:          view,
					LoadOp:        wgpu.LoadOpClear,
					StoreOp:       wgpu.StoreOpStore,
					ClearValue:    wgpu.Color{A: 1},
					ResolveTarget: nil,
				},
			},
		})

		pass.SetPipeline(g.pipeDisplay)
		pass.SetBindGroup(0, g.bindGroupDisplay, nil)
		pass.Draw(3, 1, 0, 0)
		pass.End()
	}

	buf := enc.Finish(nil)
	defer buf.Release()

	// upload a random seed value
	seeds := []uint32{rand.Uint32(), rand.Uint32()}
	ctx.WriteBuffer(g.bufSeed, 0, wgpu.ToBytes(seeds))

	ctx.Submit(buf)
}

func (g *Game) DrawToSurface(surface, offscreen *orion.Image) {
	orion.DefaultDrawToSurface(surface, offscreen, wgpu.FilterModeLinear)
	orion.DebugOverlay.Draw(surface)
}

func main() {
	err := orion.RunGame(orion.RunGameOptions{
		Game:            &Game{},
		WindowWidth:     Width,
		WindowHeight:    Height,
		WindowTitle:     "Brot",
		WindowResizable: true,
	})

	if err != nil {
		panic(err)
	}
}
