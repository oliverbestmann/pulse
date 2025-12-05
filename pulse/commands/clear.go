package commands

import (
	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/webgpu/wgpu"
)

type ClearCommand struct {
	context       *pulse.Context
	spriteCommand *SpriteCommand

	whiteTexture      *pulse.Texture
	whiteTextureClear bool
}

func NewClear(ctx *pulse.Context, spriteCommand *SpriteCommand) *ClearCommand {
	// allocate a small texture to use for clearing a sub rect of a texture
	// by simply biting the texture into the rect
	// TODO find a better solution, maybe a simpler render pipeline?
	whiteTexture := pulse.NewTexture(ctx, pulse.NewTextureOptions{
		Label:  "White",
		Format: wgpu.TextureFormatRGBA8Unorm,
		Width:  1,
		Height: 1,
	})

	return &ClearCommand{
		context:       ctx,
		spriteCommand: spriteCommand,
		whiteTexture:  whiteTexture,
	}
}

func (c *ClearCommand) Clear(target *pulse.Texture, color pulse.Color) {
	enc := c.context.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "ClearTexture"})
	defer enc.Release()

	view, resolveView := target.RenderViews()

	if target == target.Root() {
		carr := color.ToWGPU()

		desc := &wgpu.RenderPassDescriptor{
			Label: "ClearUsingTexture",
			ColorAttachments: []wgpu.RenderPassColorAttachment{
				{
					View:          view,
					ResolveTarget: resolveView,
					LoadOp:        wgpu.LoadOpClear,
					StoreOp:       wgpu.StoreOpStore,
					ClearValue: wgpu.Color{
						R: float64(carr[0]),
						G: float64(carr[1]),
						B: float64(carr[2]),
						A: float64(carr[3]),
					},
				},
			},
		}

		enc.BeginRenderPass(desc).End()

		// encode into a command buffer
		buf := enc.Finish(&wgpu.CommandBufferDescriptor{Label: "ClearTexture"})
		defer buf.Release()

		c.context.Submit(buf)
	} else {
		if !c.whiteTextureClear {
			// clear the white texture once
			c.whiteTextureClear = true
			c.Clear(c.whiteTexture, pulse.ColorWhite)
		}

		tw, th := target.Size().ToVec2f().XY()

		// draw a color square
		c.spriteCommand.Draw(target, c.whiteTexture, DrawSpriteOptions{
			Transform:    glm.ScaleMat3(tw, th),
			Color:        color,
			FilterMode:   wgpu.FilterModeNearest,
			BlendState:   wgpu.BlendStateReplace,
			AddressModeU: wgpu.AddressModeClampToEdge,
			AddressModeV: wgpu.AddressModeClampToEdge,
		})
	}
}

func (c *ClearCommand) Flush() {
	c.spriteCommand.Flush()
}
