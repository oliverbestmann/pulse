package pulse

import (
	"github.com/cogentcore/webgpu/wgpu"
)

type ClearCommand struct {
	device *wgpu.Device
}

func NewClear(ctx *Context) *ClearCommand {
	return &ClearCommand{device: ctx.Device}
}

func (c *ClearCommand) Clear(target *RenderTarget, color Color) error {
	enc, err := c.device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{
		Label: "ClearTexture",
	})

	if err != nil {
		return err
	}

	defer enc.Release()

	pass := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "ClearTexture",
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:          target.View,
				ResolveTarget: target.ResolveTarget,
				LoadOp:        wgpu.LoadOpClear,
				StoreOp:       wgpu.StoreOpStore,
				ClearValue: wgpu.Color{
					R: float64(color[0]),
					G: float64(color[1]),
					B: float64(color[2]),
					A: float64(color[3]),
				},
			},
		},
	})

	passGuard := NewReleaseGuard(pass)
	defer passGuard.Release()

	if err := pass.End(); err != nil {
		return err
	}

	passGuard.Release()

	// encode into a command buffer
	buf, err := enc.Finish(&wgpu.CommandBufferDescriptor{Label: "ClearTexture"})
	if err != nil {
		return err
	}

	defer buf.Release()

	queue := c.device.GetQueue()
	defer queue.Release()

	queue.Submit(buf)

	return nil
}

func (c *ClearCommand) Flush() error {
	return nil
}

type Releaser interface {
	Release()
}

type ReleaseGuard struct {
	delegate Releaser
}

func NewReleaseGuard(delegate Releaser) ReleaseGuard {
	return ReleaseGuard{delegate: delegate}
}

func (r *ReleaseGuard) Keep() {
	r.delegate = nil
}

func (r *ReleaseGuard) Release() {
	if r.delegate != nil {
		r.delegate.Release()
		r.delegate = nil
	}
}
