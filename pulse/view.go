package pulse

import (
	"log/slog"

	"github.com/oliverbestmann/webgpu/wgpu"
)

type View struct {
	*Context

	surfaceConfig *wgpu.SurfaceConfiguration

	// only configured if we have a multisample texture configured
	msaaTexture *Texture

	// depth texture to render to.
	// has the same sampleCount as the surface itself
	depthTexture *Texture

	sampleCount uint32

	// true if depth is enabled
	depth bool
}

func NewView(dev *Context, msaa bool, depth bool) *View {
	st := &View{Context: dev, depth: depth}

	if msaa {
		st.sampleCount = 4
	} else {
		st.sampleCount = 1
	}

	// Print the available render formats
	caps := dev.Surface.GetCapabilities(dev.Adapter)
	slog.Info("Available surface formats", slog.Any("formats", caps.Formats))

	st.surfaceConfig = &wgpu.SurfaceConfiguration{
		Usage:       wgpu.TextureUsageRenderAttachment,
		Format:      wgpu.TextureFormatBGRA8Unorm,
		PresentMode: wgpu.PresentModeFifo,
		AlphaMode:   caps.AlphaModes[0],

		// try to reduce input latency
		DesiredMaximumFrameLatency: 1,
	}

	return st
}

func (vs *View) MSAA() bool {
	return vs.sampleCount > 1
}

func (vs *View) Depth() bool {
	return vs.depth
}

func (vs *View) SurfaceAsTexture(screen *wgpu.Texture, screenView *wgpu.TextureView) *Texture {
	if vs.MSAA() {
		screenTexture := WrapTexture(screen, screenView, nil)

		return WrapTexture(
			vs.msaaTexture.texture,
			vs.msaaTexture.textureView,
			screenTexture,
		)
	} else {
		return WrapTexture(
			screen,
			screenView,
			nil,
		)
	}
}

func (vs *View) ReleaseTexture() {
	if vs.depthTexture != nil {
		vs.depthTexture.Release()
	}
	if vs.msaaTexture != nil {
		vs.msaaTexture.Release()
	}
}

func (vs *View) Configure(width, height uint32) {
	vs.surfaceConfig.Width = width
	vs.surfaceConfig.Height = height
	vs.Surface.Configure(vs.Device, vs.surfaceConfig)

	// release depth depth texture
	vs.ReleaseTexture()

	// create depth texture
	if vs.depth {
		vs.depthTexture = createDepthTexture(vs.Context, width, height, vs.sampleCount)
	}

	if vs.MSAA() {
		// create msaa render target texture
		vs.msaaTexture = createMultisampleTexture(vs.Context, vs.surfaceConfig, vs.sampleCount)
	}
}

func createMultisampleTexture(ctx *Context, surfaceConfig *wgpu.SurfaceConfiguration, sampleCount uint32) *Texture {
	return NewTextureFromDesc(ctx, &wgpu.TextureDescriptor{
		Label: "MultisampleRenderTarget",
		Usage: wgpu.TextureUsageRenderAttachment,
		Size: wgpu.Extent3D{
			Width:              surfaceConfig.Width,
			Height:             surfaceConfig.Height,
			DepthOrArrayLayers: 1,
		},
		Format:        surfaceConfig.Format,
		Dimension:     wgpu.TextureDimension2D,
		SampleCount:   sampleCount,
		MipLevelCount: 1,
	})
}

func createDepthTexture(ctx *Context, width, height, sampleCount uint32) *Texture {
	return NewTextureFromDesc(ctx, &wgpu.TextureDescriptor{
		Label:     "DepthTexture",
		Usage:     wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageTextureBinding,
		Dimension: wgpu.TextureDimension2D,
		Size: wgpu.Extent3D{
			Width:              width,
			Height:             height,
			DepthOrArrayLayers: 1,
		},
		Format:        wgpu.TextureFormatDepth32Float,
		MipLevelCount: 1,
		SampleCount:   sampleCount,
	})
}
