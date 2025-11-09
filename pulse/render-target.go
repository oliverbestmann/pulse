package pulse

import "github.com/cogentcore/webgpu/wgpu"

// RenderTarget holds all the information of something that can be rendered to.
// This is normally either an offscreen Texture or the screen.
type RenderTarget struct {
	View *wgpu.TextureView

	// In case of multisample rendering, this might hold the
	// texture the multisampled fragment is resolved to.
	ResolveTarget *wgpu.TextureView

	// Texture format of View
	Format wgpu.TextureFormat

	// Region of the View to render to
	Region Rectangle2[uint32]

	// The number of samples of the View texture
	SampleCount uint32

	TotalWidth  uint32
	TotalHeight uint32
}
