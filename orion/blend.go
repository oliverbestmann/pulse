package orion

import "github.com/oliverbestmann/webgpu/wgpu"

var blendComponentAdd = wgpu.BlendComponent{
	SrcFactor: wgpu.BlendFactorOne,
	DstFactor: wgpu.BlendFactorOne,
	Operation: wgpu.BlendOperationAdd,
}

var BlendStateAdd = wgpu.BlendState{
	Color: blendComponentAdd,
	Alpha: blendComponentAdd,
}

var BlendStateAlphaReplace = wgpu.BlendStateReplace
var BlendStateAlphaBlendingStraight = wgpu.BlendStateAlphaBlending
var BlendStateAlphaBlendingPremultiplied = wgpu.BlendStatePremultipliedAlphaBlending

// BlendStateDefault defines the default blend state. You can
// overwrite this to set a different default blend state
var BlendStateDefault = BlendStateAlphaBlendingStraight
