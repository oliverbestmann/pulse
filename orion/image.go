package orion

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse"
)

type Color = glm.Vec4f

type ColorScale struct {
	r1, g1, b1, a1 float32
}

func ColorScaleOf(color Color) ColorScale {
	return ColorScale{
		r1: color[0] - 1,
		g1: color[1] - 1,
		b1: color[2] - 1,
		a1: color[3] - 1,
	}
}

func ColorScaleRGBA(r, g, b, a float32) ColorScale {
	return ColorScaleOf(glm.Vec4f{r, g, b, a})
}

func (c *ColorScale) Scaled(vec Color) ColorScale {
	return ColorScaleOf(c.ToColor().Mul(vec))
}

func (c *ColorScale) ToColor() glm.Vec4f {
	return glm.Vec4f{
		c.r1 + 1,
		c.g1 + 1,
		c.b1 + 1,
		c.a1 + 1,
	}
}

type Image struct {
	texture *pulse.Texture
}

func asImage(texture *pulse.Texture) *Image {
	return &Image{texture: texture}
}

func (i *Image) Clear(color Color) {
	clr := clearCommand.Get()
	SwitchToCommand(clr)

	err := clr.Clear(i.texture, color)
	Handle(err, "clear image")
}

type DrawImageOptions struct {
	// Color to apply
	ColorScale ColorScale

	// Transform to apply to the image
	Transform glm.Mat3f

	// BlendState defines how to blend the image with the
	// existing framebuffer. The default is BlendStateDefault.
	BlendState wgpu.BlendState

	// FilterMode defaults to linear
	FilterMode wgpu.FilterMode
}

func (i *Image) DrawImage(source *Image, opts *DrawImageOptions) {
	if opts == nil {
		opts = &DrawImageOptions{}
	}

	var blendState = opts.BlendState
	if blendState == (wgpu.BlendState{}) {
		blendState = BlendStateDefault
	}

	var filterMode = opts.FilterMode
	if filterMode == wgpu.FilterModeUndefined {
		filterMode = wgpu.FilterModeLinear
	}

	sprites := spriteCommand.Get()
	SwitchToCommand(sprites)

	err := sprites.Draw(i.texture, source.texture, pulse.DrawSpriteOptions{
		Transform:    opts.Transform,
		Color:        opts.ColorScale.ToColor(),
		FilterMode:   filterMode,
		BlendState:   blendState,
		AddressModeU: wgpu.AddressModeClampToEdge,
		AddressModeV: wgpu.AddressModeClampToEdge,
	})

	Handle(err, "draw image")
}

func (i *Image) DrawImagesFromGPU(source *Image, buf *wgpu.Buffer, particleCount uint, opts *DrawImageOptions) {
	if opts == nil {
		opts = &DrawImageOptions{}
	}

	var blendState = opts.BlendState
	if blendState == (wgpu.BlendState{}) {
		blendState = BlendStateDefault
	}

	var filterMode = opts.FilterMode
	if filterMode == wgpu.FilterModeUndefined {
		filterMode = wgpu.FilterModeLinear
	}

	sprites := spriteCommand.Get()
	SwitchToCommand(sprites)

	err := sprites.DrawFromGPU(i.texture, source.texture, pulse.DrawSpriteFromGPUOptions{
		Buffer:        buf,
		InstanceCount: particleCount,
		FilterMode:    filterMode,
		BlendState:    blendState,
		AddressModeU:  wgpu.AddressModeClampToEdge,
		AddressModeV:  wgpu.AddressModeClampToEdge,
	})

	Handle(err, "draw image from gpu buffer")
}

func (i *Image) Size() glm.Vec2f {
	return i.texture.Sizef()
}

func (i *Image) Width() uint32 {
	return i.texture.Width()
}

func (i *Image) Height() uint32 {
	return i.texture.Height()
}

func (i *Image) Format() wgpu.TextureFormat {
	return i.texture.Format()
}

func (i *Image) MSAA() bool {
	return i.texture.MSAA()
}

func (i *Image) SubImage(x, y, width, height uint32) *Image {
	pos := glm.Vec2[uint32]{x, y}
	size := glm.Vec2[uint32]{width, height}

	subTexture := i.texture.SubTexture(pos, size)
	return &Image{
		texture: subTexture,
	}
}

func DecodeImageFromBytes(buf []byte) (*Image, error) {
	ctx := currentContext.Get()

	texture, err := pulse.DecodeTextureFromMemory(ctx, buf)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	return asImage(texture), nil
}

type NewImageOptions struct {
	// Format of the new texture. Defaults to rgba8unorm if not specified
	Format wgpu.TextureFormat

	// Enable MSAA when rendering to this texture
	MSAA bool

	// Helpful label for wgpu error messages
	Label string
}

func NewImage(width, height uint32, opts *NewImageOptions) *Image {
	if opts == nil {
		opts = &NewImageOptions{}
	}

	if opts.Format == 0 {
		opts.Format = wgpu.TextureFormatRGBA8Unorm
	}

	texture, err := pulse.NewTexture(currentContext.Get(), pulse.NewTextureOptions{
		Width:  width,
		Height: height,
		Format: opts.Format,
		MSAA:   opts.MSAA,
		Label:  opts.Label,
	})

	Handle(err, "create new texture")

	return asImage(texture)
}
