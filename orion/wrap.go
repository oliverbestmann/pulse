package orion

import (
	"github.com/cogentcore/webgpu/wgpu"
)

func CreateShaderModule(descriptor wgpu.ShaderModuleDescriptor) *wgpu.ShaderModule {
	ctx := CurrentContext()

	res, err := ctx.CreateShaderModule(&descriptor)
	Handle(err, "create shader module %q", descriptor.Label)

	return res
}

func CreateComputePipeline(descriptor wgpu.ComputePipelineDescriptor) *wgpu.ComputePipeline {
	ctx := CurrentContext()

	res, err := ctx.CreateComputePipeline(&descriptor)
	Handle(err, "create compute pipeline")

	return res
}

func CreateBindGroup(descriptor wgpu.BindGroupDescriptor) *wgpu.BindGroup {
	ctx := CurrentContext()

	res, err := ctx.CreateBindGroup(&descriptor)
	Handle(err, "create bind group %q", descriptor.Label)

	return res
}

func GetQueue() *wgpu.Queue {
	ctx := CurrentContext()
	queue := ctx.GetQueue()
	return queue
}

type CommandEncoder struct {
	*wgpu.CommandEncoder
	label string
}

func CreateCommandEncoder(label string) CommandEncoder {
	ctx := CurrentContext()

	desc := wgpu.CommandEncoderDescriptor{Label: label}
	res, err := ctx.CreateCommandEncoder(&desc)
	Handle(err, "create command encoder %q", label)

	return CommandEncoder{
		CommandEncoder: res,
		label:          label,
	}
}

func (enc *CommandEncoder) AddRenderPass(desc wgpu.RenderPassDescriptor, configure func(pass *wgpu.RenderPassEncoder)) {
	pass := enc.BeginRenderPass(&desc)
	defer pass.Release()

	// configure the render pass
	configure(pass)

	// and finalize it
	err := pass.End()
	Handle(err, "create render pass %q", desc.Label)
}

func (enc *CommandEncoder) AddComputePass(configure func(pass *wgpu.ComputePassEncoder)) {
	pass := enc.BeginComputePass(nil)
	defer pass.Release()

	// configure the render pass
	configure(pass)

	// and finalize it
	err := pass.End()
	Handle(err, "create compute pass")
}

func (enc *CommandEncoder) Finish() *wgpu.CommandBuffer {
	buf, err := enc.CommandEncoder.Finish(&wgpu.CommandBufferDescriptor{Label: enc.label})
	Handle(err, "finish command encoder %q", enc.label)

	return buf
}

func (enc *CommandEncoder) Submit() {
	Submit(enc.Finish())

}

func Submit(cmd *wgpu.CommandBuffer) wgpu.SubmissionIndex {
	ctx := CurrentContext()

	queue := ctx.GetQueue()
	defer queue.Release()

	return queue.Submit(cmd)
}
