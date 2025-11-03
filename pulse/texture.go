package pulse

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"runtime"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
)

type Texture struct {
	// point to root Texture this texture is a part of.
	root *Texture

	texture     *wgpu.Texture
	textureView *wgpu.TextureView

	resolveTarget     *wgpu.Texture
	resolveTargetView *wgpu.TextureView

	format      wgpu.TextureFormat
	sampleCount uint32

	x, y, width, height uint32
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

	var resolveTarget *wgpu.Texture
	var resolveTargetView *wgpu.TextureView

	if desc.SampleCount > 1 {
		// create resolve target texture
		descResolve := *desc
		descResolve.SampleCount = 1

		resolveTarget, err = ctx.Device.CreateTexture(&descResolve)
		if err != nil {
			textureView.Release()
			texture.Release()

			return nil, err
		}

		// now create a default texture view
		resolveTargetView, err = resolveTarget.CreateView(nil)
		if err != nil {
			textureView.Release()
			texture.Release()

			resolveTarget.Release()

			return nil, err
		}
	}

	t := &Texture{
		texture:           texture,
		textureView:       textureView,
		resolveTarget:     resolveTarget,
		resolveTargetView: resolveTargetView,

		format:      desc.Format,
		width:       desc.Size.Width,
		height:      desc.Size.Height,
		sampleCount: desc.SampleCount,
	}

	// texture itself is the root
	t.root = t

	runtime.SetFinalizer(t, func(t *Texture) {
		if t.root != t {
			return
		}

		// release wgpu texture instance
		t.textureView.Release()
		t.texture.Release()

		if t.resolveTarget != nil {
			t.resolveTarget.Release()
		}

		if t.resolveTargetView != nil {
			t.resolveTargetView.Release()
		}
	})

	return t, nil
}

func ImportTexture(texture *wgpu.Texture, textureView *wgpu.TextureView, resolveTarget *wgpu.Texture, resolveTargetView *wgpu.TextureView) *Texture {
	t := &Texture{
		texture:           texture,
		textureView:       textureView,
		resolveTarget:     resolveTarget,
		resolveTargetView: resolveTargetView,
		format:            texture.GetFormat(),
		sampleCount:       texture.GetSampleCount(),
		width:             texture.GetWidth(),
		height:            texture.GetHeight(),
	}

	t.root = t

	return t
}

func (t *Texture) Width() uint32 {
	return t.width
}

func (t *Texture) Height() uint32 {
	return t.height
}

func (t *Texture) Size() glm.Vec2f {
	return glm.Vec2f{float32(t.width), float32(t.height)}
}

func (t *Texture) UVOffset() glm.Vec2f {
	return glm.Vec2f{
		float32(t.x) / float32(t.root.width),
		float32(t.y) / float32(t.root.height),
	}
}

func (t *Texture) UVScale() glm.Vec2f {
	return glm.Vec2f{
		float32(t.width) / float32(t.root.width),
		float32(t.height) / float32(t.root.height),
	}
}

func (t *Texture) SubTexture(pos glm.Vec2[uint32], size glm.Vec2[uint32]) *Texture {
	sub := *t

	posX, posY := pos.XY()
	sub.x = t.x + posX
	sub.y = t.y + posY

	sub.width, sub.height = size.XY()

	return &sub
}

func (t *Texture) AsRenderTarget() *RenderTarget {
	return &RenderTarget{
		View:          t.textureView,
		Format:        t.format,
		Width:         t.width,
		Height:        t.height,
		SampleCount:   t.sampleCount,
		ResolveTarget: t.resolveTargetView,
	}
}

func (t *Texture) SourceView() *wgpu.TextureView {
	if t.resolveTargetView != nil {
		return t.resolveTargetView
	}

	return t.textureView
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
		BytesPerRow:  t.width * 4,
		RowsPerImage: t.height,
	}

	size := &wgpu.Extent3D{
		Width:              t.width,
		Height:             t.height,
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
