package wx

import (
	"errors"
	"log/slog"

	"github.com/oliverbestmann/webgpu/wgpu"
)

var ErrSurfaceNotConfigured = errors.New("surface not configured")

type Surface struct {
	*Context

	surfaceConfig *wgpu.SurfaceConfiguration

	// only set if we have a multisample texture configured
	msaaTexture *Texture

	// depth texture to render to.
	// has the same sampleCount as the surface itself
	depthTexture *Texture

	sampleCount uint32

	// true if depth is enabled
	depth bool

	// true if configured was called
	configured bool
}

func NewSurface(ctx *Context, msaa bool, depth bool) *Surface {
	st := &Surface{Context: ctx, depth: depth}

	if msaa {
		st.sampleCount = 4
	} else {
		st.sampleCount = 1
	}

	// Print the available render formats
	caps := ctx.Surface.GetCapabilities(ctx.Adapter)
	slog.Info("Available surface formats", slog.Any("formats", caps.Formats))

	st.surfaceConfig = &wgpu.SurfaceConfiguration{
		Usage:       wgpu.TextureUsageRenderAttachment,
		Format:      wgpu.TextureFormatBGRA8Unorm,
		PresentMode: wgpu.PresentModeFifo,
		AlphaMode:   caps.AlphaModes[0],
		ViewFormats: []wgpu.TextureFormat{
			wgpu.TextureFormatBGRA8UnormSrgb,
		},

		// try to reduce input latency
		DesiredMaximumFrameLatency: 1,
	}

	return st
}

func (s *Surface) MSAA() bool {
	return s.sampleCount > 1
}

func (s *Surface) Depth() *Texture {
	if !s.configured {
		panic(ErrSurfaceNotConfigured)
	}

	return s.depthTexture
}

func (s *Surface) AsTexture(screen *wgpu.Texture, screenView *wgpu.TextureView) *Texture {
	if !s.configured {
		panic(ErrSurfaceNotConfigured)
	}

	screenTexture := WrapTexture(screen, WrapTextureOptions{
		TextureView:       screenView,
		TextureViewFormat: wgpu.TextureFormatBGRA8UnormSrgb,
	})

	if s.MSAA() {
		return WrapTexture(s.msaaTexture.texture, WrapTextureOptions{
			TextureViewFormat: wgpu.TextureFormatBGRA8UnormSrgb,
			TextureView:       s.msaaTexture.textureView,
			ResolveTarget:     screenTexture,
		})
	}

	return screenTexture
}

func (s *Surface) Configure(width, height uint32) {
	s.configured = true
	s.surfaceConfig.Width = width
	s.surfaceConfig.Height = height
	s.Surface.Configure(s.Device, s.surfaceConfig)

	// release depth depth texture
	s.ReleaseTextures()

	// create depth texture
	if s.depth {
		s.depthTexture = createDepthTexture(s.Context, width, height, s.sampleCount)
	}

	if s.MSAA() {
		// create msaa render target texture
		s.msaaTexture = createMSAATexture(s.Context, s.surfaceConfig, s.sampleCount)
	}
}

func (s *Surface) ReleaseTextures() {
	s.configured = false

	if s.depthTexture != nil {
		s.depthTexture.Release()
	}

	if s.msaaTexture != nil {
		s.msaaTexture.Release()
	}
}

func createMSAATexture(ctx *Context, surfaceConfig *wgpu.SurfaceConfiguration, sampleCount uint32) *Texture {
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
