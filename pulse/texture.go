package pulse

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"

	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/webgpu/wgpu"
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

func NewTexture(ctx *Context, opts NewTextureOptions) *Texture {
	var sampleCount uint32 = 1

	if isOpenGL(ctx) {
		// https://github.com/gfx-rs/wgpu/issues/6084
		// In OpenGL backend, attempting to render to a multisampled texture
		// with the usages RENDER_ATTACHMENT | TEXTURE_BINDING fails
		opts.MSAA = false
	}

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
func NewTextureFromDesc(ctx *Context, desc *wgpu.TextureDescriptor) *Texture {
	texture := ctx.Device.CreateTexture(desc)

	// now create a default texture view
	textureView := texture.CreateView(nil)

	var resolveTarget *Texture

	if desc.SampleCount > 1 {
		// create resolve target texture
		descResolve := *desc
		descResolve.SampleCount = 1

		resolveTarget = NewTextureFromDesc(ctx, &descResolve)
	}

	region := RectangleFromSize(
		glm.Vec2u{},
		glm.Vec2u{
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

	return t
}

// WrapTexture creates a texture from an existing wgpu.Texture and wgpu.TextureView. If it is a
// multisample texture, you also need to specify a resolve target.
func WrapTexture(texture *wgpu.Texture, textureView *wgpu.TextureView, resolveTarget *Texture) *Texture {
	if texture.GetSampleCount() > 1 && resolveTarget == nil {
		panic("no resolveTarget specified for multisample texture")
	}

	if texture.GetSampleCount() == 1 && resolveTarget != nil {
		panic("resolveTarget specified for multisample texture")
	}

	region := RectangleFromSize(
		glm.Vec2u{},
		glm.Vec2u{
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

func (t *Texture) SubTexture(pos glm.Vec2u, size glm.Vec2u) *Texture {
	sub := *t

	pos = t.region.Min.Add(pos)
	sub.region = RectangleFromSize(pos, size)

	if sub.resolveTarget != nil {
		sub.resolveTarget = sub.resolveTarget.SubTexture(pos, size)
	}

	return &sub
}

func (t *Texture) Root() *Texture {
	return t.root
}

func (t *Texture) IsSubTexture() bool {
	return t != t.root
}

func (t *Texture) Width() uint32 {
	return t.region.Width()
}

func (t *Texture) Height() uint32 {
	return t.region.Height()
}

func (t *Texture) Offset() glm.Vec2u {
	return t.region.Offset()
}

func (t *Texture) Size() glm.Vec2u {
	return t.region.Size()
}

func (t *Texture) UV() Rectangle2f {
	rootSize := t.root.Size().ToVec2f()
	uvOffset := t.Offset().ToVec2f().Div(rootSize)
	uvScale := t.Size().ToVec2f().Div(rootSize)
	return RectangleFromSize(uvOffset, uvScale)
}

func (t *Texture) Format() wgpu.TextureFormat {
	return t.format
}

func (t *Texture) SampleCount() uint32 {
	return t.sampleCount
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

func (t *Texture) SourceView() *wgpu.TextureView {
	if t.resolveTarget != nil {
		return t.resolveTarget.textureView
	}

	return t.textureView
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

func (t *Texture) RenderViews() (view, resolveView *wgpu.TextureView) {
	view = t.textureView

	if t.sampleCount > 1 {
		resolveView = t.resolveTarget.textureView
	}

	return
}

func (t *Texture) WritePixels(ctx *Context, pixels []byte) {
	rect := RectangleFromXYWH(0, 0, t.Width(), t.Height())

	t.WritePixelsToRect(ctx, WritePixelsOptions{
		Pixels: pixels,
		Region: rect,
	})
}

type WritePixelsOptions struct {
	Pixels   []byte
	Region   Rectangle2u
	Stride   uint32
	MipLevel uint32
}

func (t *Texture) WritePixelsToRect(ctx *Context, opts WritePixelsOptions) {
	// fail if not in rect
	if !t.region.Contains(opts.Region) {
		return
	}

	if opts.Stride == 0 {
		opts.Stride = opts.Region.Width() * 4
	}

	layout := &wgpu.TexelCopyBufferLayout{
		Offset:       0,
		BytesPerRow:  opts.Stride,
		RowsPerImage: opts.Region.Height(),
	}

	size := &wgpu.Extent3D{
		Width:              opts.Region.Width(),
		Height:             opts.Region.Height(),
		DepthOrArrayLayers: 1,
	}

	dest := &wgpu.TexelCopyTextureInfo{
		Texture:  t.texture,
		MipLevel: opts.MipLevel,
		Origin: wgpu.Origin3D{
			X: opts.Region.Min[0],
			Y: opts.Region.Min[1],
		},
		Aspect: wgpu.TextureAspectAll,
	}

	// send data to the gpu
	ctx.WriteTexture(dest, opts.Pixels, layout, size)
}

func DecodeTextureFromMemory(ctx *Context, buf []byte) (*Texture, error) {
	src, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("decode image from memory: %w", err)
	}

	tex := NewTextureFromImage(ctx, src)
	return tex, nil
}

func NewTextureFromImage(ctx *Context, src image.Image) *Texture {
	iw, ih := src.Bounds().Dx(), src.Bounds().Dy()
	rgba := image.NewNRGBA(image.Rect(0, 0, iw, ih))

	draw.Draw(rgba, rgba.Bounds(), src, image.Point{}, draw.Src)

	t := NewTexture(ctx, NewTextureOptions{
		// TODO handle srgb import too
		Format: wgpu.TextureFormatRGBA8Unorm,
		Width:  uint32(iw),
		Height: uint32(ih),
		Label:  "",
	})

	t.WritePixels(ctx, rgba.Pix)

	return t
}

func isOpenGL(ctx *Context) bool {
	return ctx.Adapter.GetInfo().BackendType == wgpu.BackendTypeOpenGL
}
