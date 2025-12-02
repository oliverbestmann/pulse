package main

import (
	_ "embed"
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/orion"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed gopher.jpg
var gopherImage []byte

//go:embed particle.png
var particleImage []byte

//go:embed particles.wgsl
var particlesShader string

type Particle struct {
	Transform glm.Mat3f
	Color     orion.Color
	Velocity  glm.Vec2f
}

type TestGame struct {
	orion.DefaultGame

	lastTime time.Time

	time float32

	gopher     *orion.Image
	tempTarget *orion.Image
	particle   *orion.Image

	particles       []Particle
	particleCommand *ParticleCommand
	iconScale       float32
}

func (g *TestGame) Layout(surfaceWidth, surfaceHeight uint32) orion.LayoutOptions {
	return orion.LayoutOptions{
		Width:  1000,
		Height: 600,
	}
}

func (g *TestGame) Initialize() error {
	gopher, err := orion.DecodeImageFromBytes(gopherImage)
	if err != nil {
		return fmt.Errorf("decode gopher texture: %w", err)
	}

	particle, err := orion.DecodeImageFromBytes(particleImage)
	if err != nil {
		return fmt.Errorf("decode particle texture: %w", err)
	}

	tempTarget := orion.NewImage(1000, 600, &orion.NewImageOptions{
		Label:  "TempTarget",
		Format: wgpu.TextureFormatRGBA16Float,
		MSAA:   true,
	})

	g.gopher = gopher
	g.particle = particle
	g.tempTarget = tempTarget

	rng := rand.New(rand.NewPCG(1, 2))

	for range 100_000 {
		x := rng.Float32()*800 + 100
		y := rng.Float32()*400 + 100
		scale := (rng.Float32()*rng.Float32())*0.8 + 0.2

		cb := randf(rng, 0.7, 0.9)
		cr := randf(rng, 0.5, 0.8) * cb
		cg := randf(rng, 0.5, 0.8) * cb
		color := orion.Color{cr, cg, cb, 1}.Scale(0.05)

		transform := glm.TranslationMat3(x, y).
			Scale(scale, scale).
			Translate(-32, -32)

		velocity := glm.Vec2f{
			randf(rng, -16.0, 16.0),
			randf(rng, -16.0, 16.0),
		}

		g.particles = append(g.particles, Particle{
			Transform: transform,
			Color:     color,
			Velocity:  velocity,
		})
	}

	particleCommand := NewParticleCommand(g.particles)

	g.iconScale = 1.0
	g.particleCommand = particleCommand

	orion.DebugOverlay.Enable(true)
	orion.DebugOverlay.RunGC = true

	return nil
}

func (g *TestGame) Update() error {
	now := time.Now()
	if g.lastTime.IsZero() {
		g.lastTime = now
	}

	dt := float32(now.Sub(g.lastTime).Seconds())
	g.lastTime = now

	g.time += dt
	g.particleCommand.Execute(dt)

	if orion.IsKeyPressed(32) {
		g.iconScale += 10 * dt
	}

	g.iconScale = max(1.0, g.iconScale-dt)

	return nil
}

func (g *TestGame) Draw(screen *orion.Image) {
	t := g.time * 10

	// clear the screen texture
	screen.Clear(orion.Color{0.2, 0.2, 0.3, 1.0})

	// clear our temporary test texture
	g.tempTarget.Clear(orion.Color{0.2, 0.3, 0.2, 0.1})

	mouse := orion.MousePosition()

	g.tempTarget.DrawImage(g.gopher, &orion.DrawImageOptions{
		Transform: glm.TranslationMat3(mouse.XY()).
			Rotate(glm.Rad(t*0.2)).
			Scale(g.iconScale, g.iconScale).
			Scale(0.25, 0.25).
			Translate(g.gopher.Sizef().Scale(-0.5).XY()),
	})

	// render particles directly from the gpu buffer
	particleCount := uint(len(g.particles))
	screen.DrawImagesFromGPU(g.particle, g.particleCommand.bufParticlesSprites, particleCount, &orion.DrawImageOptions{
		BlendState: orion.BlendStateAdd,
	})

	// draw to screen
	screen.DrawImage(g.tempTarget, nil)

	orion.DebugOverlay.Draw(screen)
}

func randf(rng *rand.Rand, min, max float32) float32 {
	return rng.Float32()*(max-min) + min
}

type gpuParticle struct {
	Transform [12]float32
	Color     [4]float32
	Velocity  [2]float32

	// padding
	_ [8]byte
}

type ParticleCommand struct {
	particles           []gpuParticle
	bufParticlesGPU     *wgpu.Buffer
	bufParticlesSprites *wgpu.Buffer
	bufTime             *wgpu.Buffer
	pipeline            *wgpu.ComputePipeline
	bindGroup           *wgpu.BindGroup

	// set to true while data is currently being mapped into cpu memory.
	// the bufParticlesCPU buffer must not be queued during that time
	reading atomic.Bool
}

func NewParticleCommand(particlesIn []Particle) *ParticleCommand {
	if unsafe.Sizeof(gpuParticle{}) != 80 {
		panic("gpu particle has wrong size")
	}

	dev := orion.CurrentContext()

	var particles []gpuParticle
	for _, particle := range particlesIn {
		particles = append(particles, gpuParticle{
			Transform: particle.Transform.ToWGPU(),
			Color:     particle.Color.ToWGPU(),
			Velocity:  particle.Velocity.ToWGPU(),
		})
	}

	bufParticlesGPU := dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Particles.GPU",
		Size:  uint64(len(wgpu.ToBytes(particles))),
		Usage: wgpu.BufferUsageStorage | wgpu.BufferUsageCopyDst | wgpu.BufferUsageCopySrc,
	})

	// staging buffer buffer
	bufParticlesSprites := dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Particles.Sprites",
		Size:  uint64((4 + 2 + 2 + 3 + 3) * 4 * len(particles)),
		Usage: wgpu.BufferUsageStorage | wgpu.BufferUsageVertex,
	})

	// delta time buffer
	bufTime := dev.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label:    "DeltaTime",
		Contents: wgpu.ToBytes([]float32{float32(0)}),
		Usage:    wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})

	// build the pipeline
	shader := dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:      "ParticlesComputeShader",
		WGSLSource: &wgpu.ShaderSourceWGSL{Code: particlesShader},
	})

	// create the pipeline
	pipeline := dev.CreateComputePipeline(&wgpu.ComputePipelineDescriptor{
		Compute: wgpu.ProgrammableStageDescriptor{
			Module:     shader,
			EntryPoint: "update_particles",
		},
	})

	// need to get the layout
	bindGroupLayout := pipeline.GetBindGroupLayout(0)
	defer bindGroupLayout.Release()

	// create the bind group
	bindGroup := dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "ParticleBindGroup",
		Layout: bindGroupLayout,
		Entries: []wgpu.BindGroupEntry{
			{
				Buffer:  bufParticlesGPU,
				Size:    wgpu.WholeSize,
				Binding: 0,
			},
			{
				Buffer:  bufParticlesSprites,
				Size:    wgpu.WholeSize,
				Binding: 1,
			},
			{
				Buffer:  bufTime,
				Size:    4,
				Binding: 2,
			},
		},
	})

	// copy data to the gpu
	orion.WriteSliceToBuffer(bufParticlesGPU, wgpu.ToBytes(particles))

	return &ParticleCommand{
		particles:           particles,
		bufParticlesGPU:     bufParticlesGPU,
		bufParticlesSprites: bufParticlesSprites,
		bufTime:             bufTime,
		pipeline:            pipeline,
		bindGroup:           bindGroup,
	}
}

func (p *ParticleCommand) Execute(dt float32) {
	orion.WriteValueToBuffer(p.bufTime, dt)

	count := uint32(len(p.particles))
	xSize := count / 64
	ySize := uint32(64)

	dev := orion.CurrentContext()

	encoder := dev.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "UpdateParticlesEnc"})

	pass := encoder.BeginComputePass(&wgpu.ComputePassDescriptor{Label: "UpdateParticles"})
	pass.SetPipeline(p.pipeline)
	pass.SetBindGroup(0, p.bindGroup, nil)
	pass.DispatchWorkgroups(xSize, ySize, 1)
	pass.End()

	// submit commands to queue
	dev.Submit(encoder.Finish(nil))
}
