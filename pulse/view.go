package pulse

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
)

type View struct {
	*Context

	surfaceConfig *wgpu.SurfaceConfiguration

	sampleCount uint32

	// only configured if we have a multisample texture configured
	msaaTexture *Texture

	// depth texture to render to.
	// has the same sampleCount as the surface itself
	depthTexture *Texture
}

func NewView(dev *Context, msaa bool) (st *View, err error) {
	defer func() {
		if err != nil && st != nil {
			st.Release()
			st = nil
		}
	}()

	st = &View{Context: dev}

	if msaa {
		st.sampleCount = 4
	} else {
		st.sampleCount = 1
	}

	// Print the available render formats
	caps := dev.Surface.GetCapabilities(dev.Adapter)
	fmt.Println("Available surface formats", caps.Formats)

	st.surfaceConfig = &wgpu.SurfaceConfiguration{
		Usage:       wgpu.TextureUsageRenderAttachment,
		Format:      wgpu.TextureFormatBGRA8Unorm,
		PresentMode: wgpu.PresentModeFifo,
		AlphaMode:   caps.AlphaModes[0],

		// try to reduce input latency
		DesiredMaximumFrameLatency: 1,
	}

	return st, nil
}

func (vs *View) MSAA() bool {
	return vs.sampleCount > 1
}

func (vs *View) AsTexture(screen *wgpu.Texture, screenView *wgpu.TextureView) *Texture {
	// vs.surfaceConfig.Format
	// vs.surfaceConfig.Width
	// vs.surfaceConfig.Height

	if vs.MSAA() {
		screenTexture := ImportTexture(screen, screenView, nil)

		return ImportTexture(
			vs.msaaTexture.texture,
			vs.msaaTexture.textureView,
			screenTexture,
		)
	} else {
		return ImportTexture(
			screen,
			screenView,
			nil,
		)
	}
}

func (vs *View) Release() {
	if vs.depthTexture != nil {
		vs.depthTexture.Release()
	}
	vs.depthTexture = nil
	vs.msaaTexture = nil
}

func (vs *View) Configure(width, height uint32) error {
	vs.surfaceConfig.Width = width
	vs.surfaceConfig.Height = height
	vs.Surface.Configure(vs.Device, vs.surfaceConfig)

	var err error

	// release previous textures
	vs.depthTexture = nil
	vs.msaaTexture = nil

	// create depth texture
	vs.depthTexture, err = createDepthTexture(vs.Context, width, height, vs.sampleCount)
	if err != nil {
		return err
	}

	if vs.MSAA() {
		// create msaa render target texture
		vs.msaaTexture, err = createMultisampleTexture(vs.Context, vs.surfaceConfig, vs.sampleCount)
		if err != nil {
			return err
		}
	}

	return nil
}

func createMultisampleTexture(ctx *Context, surfaceConfig *wgpu.SurfaceConfiguration, sampleCount uint32) (*Texture, error) {
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

func createDepthTexture(ctx *Context, width, height, sampleCount uint32) (*Texture, error) {
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
