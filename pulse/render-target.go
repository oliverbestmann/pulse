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

	// TODO we would need the offset here too i guess

	// Size of the target to render to
	Width  uint32
	Height uint32

	// The number of samples of the View texture
	SampleCount uint32
}
