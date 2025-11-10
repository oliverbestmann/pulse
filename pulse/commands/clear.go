package commands

import (
	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse"
)

type ClearCommand struct {
	device        *wgpu.Device
	spriteCommand *SpriteCommand

	whiteTexture      *pulse.Texture
	whiteTextureClear bool
}

func NewClear(ctx *pulse.Context, spriteCommand *SpriteCommand) *ClearCommand {
	// TODO find a better solution, maybe a simpler render pipeline?
	whiteTexture, _ := pulse.NewTexture(ctx, pulse.NewTextureOptions{
		Label:  "White",
		Format: wgpu.TextureFormatRGBA8Unorm,
		Width:  1,
		Height: 1,
	})

	return &ClearCommand{device: ctx.Device, spriteCommand: spriteCommand, whiteTexture: whiteTexture}
}

func (c *ClearCommand) Clear(target *pulse.Texture, color pulse.Color) error {
	enc, err := c.device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{
		Label: "ClearTexture",
	})

	if err != nil {
		return err
	}

	defer enc.Release()

	view, resolveView := target.Views()

	if target == target.Root() {
		desc := &wgpu.RenderPassDescriptor{
			Label: "ClearTexture",
			ColorAttachments: []wgpu.RenderPassColorAttachment{
				{
					View:          view,
					ResolveTarget: resolveView,
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
		}

		pass := enc.BeginRenderPass(desc)

		defer func() {
			if pass != nil {
				pass.Release()
			}
		}()

		if err := pass.End(); err != nil {
			return err
		}

		pass.Release()
		pass = nil

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
	} else {
		if !c.whiteTextureClear {
			c.whiteTextureClear = true

			if err := c.Clear(c.whiteTexture, pulse.ColorWhite); err != nil {
				return err
			}
		}

		tw, th := target.Size().ToVec2f().XY()

		// draw a color square
		return c.spriteCommand.Draw(target, c.whiteTexture, DrawSpriteOptions{
			Transform:    glm.ScaleMat3(tw, th),
			Color:        color,
			FilterMode:   wgpu.FilterModeNearest,
			BlendState:   wgpu.BlendStateReplace,
			AddressModeU: wgpu.AddressModeClampToEdge,
			AddressModeV: wgpu.AddressModeClampToEdge,
		})
	}
}

func (c *ClearCommand) Flush() error {
	return c.spriteCommand.Flush()
}
