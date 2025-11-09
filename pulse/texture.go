package pulse

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
)

// Texture wraps a wgpu.Texture and an identity wgpu.TextureView.
// For multisample textures a Texture also holds the resolve target
// texture. A Texture can represent a sub region of another texture.
type Texture struct {
	// point to root Texture this texture is a part of.
	root *Texture

	texture     *wgpu.Texture
	textureView *wgpu.TextureView

	resolveTarget *Texture

	// equal to texture.GetFormat()
	format wgpu.TextureFormat

	// equal to texture.GetSampleCount()
	sampleCount uint32

	// sub texture
	region Rectangle2u
}

type NewTextureOptions struct {
	Format wgpu.TextureFormat
	Width  uint32
	Height uint32

	MSAA  bool
	Label string
}

func NewTexture(ctx *Context, opts NewTextureOptions) (*Texture, error) {
	var sampleCount uint32 = 1

	if opts.MSAA {
		sampleCount = 4
	}

	desc := &wgpu.TextureDescriptor{
		Label:         opts.Label,
		Format:        opts.Format,
		SampleCount:   sampleCount,
		MipLevelCount: 1,

		Dimension: wgpu.TextureDimension2D,
		Size: wgpu.Extent3D{
			Width:              opts.Width,
			Height:             opts.Height,
			DepthOrArrayLayers: 1,
		},

		// allow to do almost everything with this texture
		Usage: wgpu.TextureUsageTextureBinding |
			wgpu.TextureUsageRenderAttachment |
			wgpu.TextureUsageCopyDst |
			wgpu.TextureUsageCopySrc,
	}

	return NewTextureFromDesc(ctx, desc)
}

// NewTextureFromDesc gives you full control and creates a texture directly from
// a texture descriptor
func NewTextureFromDesc(ctx *Context, desc *wgpu.TextureDescriptor) (*Texture, error) {
	texture, err := ctx.Device.CreateTexture(desc)
	if err != nil {
		return nil, err
	}

	// now create a default texture view
	textureView, err := texture.CreateView(nil)
	if err != nil {
		texture.Release()

		return nil, err
	}

	var resolveTarget *Texture

	if desc.SampleCount > 1 {
		// create resolve target texture
		descResolve := *desc
		descResolve.SampleCount = 1

		resolveTarget, err = NewTextureFromDesc(ctx, &descResolve)
		if err != nil {
			textureView.Release()
			texture.Release()

			return nil, fmt.Errorf("create resolveTarget texture: %w", err)
		}
	}

	region := RectangleFromSize(
		glm.Vec2[uint32]{},
		glm.Vec2[uint32]{
			desc.Size.Width,
			desc.Size.Height,
		},
	)

	t := &Texture{
		texture:       texture,
		textureView:   textureView,
		resolveTarget: resolveTarget,

		format:      desc.Format,
		region:      region,
		sampleCount: desc.SampleCount,
	}

	// texture itself is the root
	t.root = t

	return t, nil
}

// ImportTexture creates a texture from an existing wgpu.Texture and wgpu.TextureView. If it is a
// multisample texture, you also need to specify a resolve target.
func ImportTexture(texture *wgpu.Texture, textureView *wgpu.TextureView, resolveTarget *Texture) *Texture {
	if texture.GetSampleCount() > 1 && resolveTarget == nil {
		panic("no resolveTarget specified for multisample texture")
	}

	if texture.GetSampleCount() == 1 && resolveTarget != nil {
		panic("resolveTarget specified for multisample texture")
	}

	region := RectangleFromSize(
		glm.Vec2[uint32]{},
		glm.Vec2[uint32]{
			texture.GetWidth(),
			texture.GetHeight(),
		},
	)

	t := &Texture{
		texture:       texture,
		textureView:   textureView,
		resolveTarget: resolveTarget,
		format:        texture.GetFormat(),
		sampleCount:   texture.GetSampleCount(),
		region:        region,
	}

	t.root = t

	return t
}

func (t *Texture) SubTexture(pos glm.Vec2[uint32], size glm.Vec2[uint32]) *Texture {
	sub := *t

	pos = t.region.Min.Add(pos)
	sub.region = RectangleFromSize(pos, size)

	return &sub
}

func (t *Texture) SourceView() *wgpu.TextureView {
	if t.resolveTarget != nil {
		return t.resolveTarget.textureView
	}

	return t.textureView
}

func (t *Texture) Root() *Texture {
	return t.root
}

func (t *Texture) Width() uint32 {
	return t.region.Width()
}

func (t *Texture) Height() uint32 {
	return t.region.Height()
}

func (t *Texture) Size() glm.Vec2[uint32] {
	return t.region.Size()
}

// Widthf is a convenience function that returns Width() as float32
func (t *Texture) Widthf() float32 {
	return float32(t.region.Width())
}

// Heightf is a convenience function that returns Height() as float32
func (t *Texture) Heightf() float32 {
	return float32(t.region.Height())
}

// Sizef is a convenience function that returns Size() as float32
func (t *Texture) Sizef() glm.Vec2f {
	x, y := t.Size().XY()
	return glm.Vec2f{float32(x), float32(y)}
}

func (t *Texture) UV() Rectangle2f {
	rootSize := t.root.Sizef()
	pos := glm.Vec2f{float32(t.region.Min[0]), float32(t.region.Min[1])}
	uvOffset := pos.Div(rootSize)
	uvSize := t.Sizef().Div(rootSize)
	return RectangleFromSize(uvOffset, uvSize)
}

func (t *Texture) Format() wgpu.TextureFormat {
	return t.texture.GetFormat()
}

func (t *Texture) MSAA() bool {
	return t.texture.GetSampleCount() > 1
}

func (t *Texture) ToWGPUTexture() *wgpu.Texture {
	return t.texture
}

func (t *Texture) ToWGPUTextureView() *wgpu.TextureView {
	return t.textureView
}

func (t *Texture) ResolveTarget() *Texture {
	return t.resolveTarget
}

// Release releases the texture view. This only works for the root texture,
// not for a sub texture. You must be sure to not use the texture after
// calling release. It might be better to not call Release at all and let the
// garbage collector handle cleanup.
func (t *Texture) Release() {
	if t.root == t {
		t.textureView.Release()
		t.texture.Release()
	}
}

func (t *Texture) Views() (view, resolveView *wgpu.TextureView) {
	view = t.textureView

	if t.sampleCount > 1 {
		resolveView = t.resolveTarget.textureView
	}

	return
}

func DecodeTextureFromMemory(ctx *Context, buf []byte) (*Texture, error) {
	src, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("decode image from memory: %w", err)
	}

	return NewTextureFromImage(ctx, src)
}

func NewTextureFromImage(ctx *Context, src image.Image) (*Texture, error) {
	iw, ih := src.Bounds().Dx(), src.Bounds().Dy()
	rgba := image.NewRGBA(image.Rect(0, 0, iw, ih))

	draw.Draw(rgba, rgba.Bounds(), src, image.Point{}, draw.Src)

	t, err := NewTexture(ctx, NewTextureOptions{
		// TODO handle srgb import
		Format: wgpu.TextureFormatRGBA8Unorm,
		Width:  uint32(iw),
		Height: uint32(ih),
		Label:  "",
	})
	if err != nil {
		return nil, fmt.Errorf("create texture: %w", err)
	}

	layout := &wgpu.TexelCopyBufferLayout{
		Offset:       0,
		BytesPerRow:  t.Width() * 4,
		RowsPerImage: t.Height(),
	}

	size := &wgpu.Extent3D{
		Width:              t.Width(),
		Height:             t.Height(),
		DepthOrArrayLayers: 1,
	}

	dest := t.texture.AsImageCopy()

	queue := ctx.Device.GetQueue()
	defer queue.Release()

	err = queue.WriteTexture(dest, rgba.Pix, layout, size)
	if err != nil {
		return nil, fmt.Errorf("copy image data to texture: %w", err)
	}

	return t, nil
}
