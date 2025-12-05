package orion

import (
	"fmt"

	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/pulse"
	"github.com/oliverbestmann/pulse/pulse/commands"
	"github.com/oliverbestmann/webgpu/wgpu"
)

type Color = pulse.Color

type ColorScale = pulse.Color

type Image struct {
	texture *pulse.Texture
}

func asImage(texture *pulse.Texture) *Image {
	return &Image{texture: texture}
}

func (i *Image) Clear(color Color) {
	clr := clearCommand.Get()
	SwitchToCommand(clr)

	clr.Clear(i.texture, color)
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

	sprites.Draw(i.texture, source.texture, commands.DrawSpriteOptions{
		Transform:    opts.Transform,
		Color:        opts.ColorScale,
		FilterMode:   filterMode,
		BlendState:   blendState,
		AddressModeU: wgpu.AddressModeClampToEdge,
		AddressModeV: wgpu.AddressModeClampToEdge,
	})
}

func (i *Image) DrawImagesFromGPU(source *Image, buf *wgpu.Buffer, count uint, opts *DrawImageOptions) {
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

	sprites.DrawFromGPU(i.texture, source.texture, commands.DrawSpriteFromGPUOptions{
		Buffer:        buf,
		InstanceCount: count,
		FilterMode:    filterMode,
		BlendState:    blendState,
		AddressModeU:  wgpu.AddressModeClampToEdge,
		AddressModeV:  wgpu.AddressModeClampToEdge,
	})
}

type Vertex2d struct {
	Position glm.Vec2f
	Color    Color
}

type DrawTrianglesOptions struct {
	Transform  glm.Mat3f
	ColorScale Color
	BlendState wgpu.BlendState
	Shader     string
}

func (i *Image) DrawTriangles(vertices []Vertex2d, opts *DrawTrianglesOptions) {
	mesh := mesh2dCommand.Get()
	SwitchToCommand(mesh)

	if opts == nil {
		opts = &DrawTrianglesOptions{}
	}

	blendState := BlendStateDefault

	if opts.BlendState != (wgpu.BlendState{}) {
		blendState = opts.BlendState
	}

	transformed := make([]commands.MeshVertex, len(vertices))

	for idx := range vertices {
		transformed[idx] = commands.MeshVertex{
			Position: vertices[idx].Position,
			Color:    vertices[idx].Color.ToVec(),
		}
	}

	mesh.DrawTriangles(i.texture, commands.DrawMesh2dOptions{
		Transform:  opts.Transform,
		BlendState: blendState,
		Vertices:   transformed,
		Color:      opts.ColorScale.ToVec(),
		Shader:     opts.Shader,
	})
}

func (i *Image) Sizef() glm.Vec2f {
	return i.texture.Size().ToVec2f()
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
	return i.texture.SampleCount() > 1
}

func (i *Image) SubImage(x, y, width, height uint32) *Image {
	pos := glm.Vec2[uint32]{x, y}
	size := glm.Vec2[uint32]{width, height}

	subTexture := i.texture.SubTexture(pos, size)
	return &Image{
		texture: subTexture,
	}
}

func (i *Image) WritePixels(pixels []byte) {
	i.texture.WritePixels(CurrentContext(), pixels)
}

func (i *Image) Texture() *pulse.Texture {
	return i.texture
}

type DecodeImageOptions struct {
	// LinearRGBA indicates that the image data is to be interpreted as linear rgba.
	// This is useful if the color values are representing data, e.g. a normal map or a height map.
	// The default for image files like jpeg or png is non linear srgb.
	LinearRGBA bool
}

func DecodeImageFromBytes(buf []byte, opts *DecodeImageOptions) (*Image, error) {
	srgb := opts == nil || !opts.LinearRGBA

	ctx := currentContext.Get()

	texture, err := pulse.DecodeTextureFromMemory(ctx, buf, srgb)
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

	texture := pulse.NewTexture(currentContext.Get(), pulse.NewTextureOptions{
		Width:  width,
		Height: height,
		Format: opts.Format,
		MSAA:   opts.MSAA,
		Label:  opts.Label,
	})

	return asImage(texture)
}
