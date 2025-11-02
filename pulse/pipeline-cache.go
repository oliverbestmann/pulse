package pulse

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/hashicorp/golang-lru/v2"
)

type PipelineConfig interface {
	comparable

	// Specialize creates a specialized pipeline for the
	// current PipelineConfig
	Specialize(def *wgpu.Device) (*wgpu.RenderPipeline, error)
}

type PipelineCache[C PipelineConfig] struct {
	device *wgpu.Device
	cache  *lru.Cache[C, *wgpu.RenderPipeline]
}

func NewPipelineCache[C PipelineConfig](ctx *Context) *PipelineCache[C] {
	cache, _ := lru.NewWithEvict[C, *wgpu.RenderPipeline](16, releaseOnEviction[C])

	return &PipelineCache[C]{
		device: ctx.Device,
		cache:  cache,
	}
}

func (p *PipelineCache[C]) Get(conf C) (*wgpu.RenderPipeline, error) {
	pipeline, ok := p.cache.Get(conf)
	if ok {
		return pipeline, nil
	}

	pipeline, err := conf.Specialize(p.device)
	if err != nil {
		return nil, fmt.Errorf("build pipeline: %w", err)
	}

	p.cache.Add(conf, pipeline)

	return pipeline, nil
}

func releaseOnEviction[C any](_config C, pipe *wgpu.RenderPipeline) {
	pipe.Release()
}
