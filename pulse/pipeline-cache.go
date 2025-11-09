package pulse

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/hashicorp/golang-lru/v2"
)

type CachedPipeline struct {
	Pipeline   *wgpu.RenderPipeline
	bindGroups *lru.Cache[uint32, *wgpu.BindGroupLayout]
}

func (pc *CachedPipeline) GetBindGroupLayout(idx uint32) *wgpu.BindGroupLayout {
	bindGroup, ok := pc.bindGroups.Get(idx)
	if ok {
		return bindGroup
	}

	bindGroup = pc.Pipeline.GetBindGroupLayout(idx)
	pc.bindGroups.Add(idx, bindGroup)

	return bindGroup
}

type PipelineConfig interface {
	comparable

	// Specialize creates a specialized pipeline for the
	// current PipelineConfig
	Specialize(def *wgpu.Device) (*wgpu.RenderPipeline, error)
}

type PipelineCache[C PipelineConfig] struct {
	device *wgpu.Device
	cache  *lru.Cache[C, CachedPipeline]
}

func NewPipelineCache[C PipelineConfig](ctx *Context) *PipelineCache[C] {
	cache, _ := lru.NewWithEvict[C, CachedPipeline](16, releasePipelineOnEviction[C])

	return &PipelineCache[C]{
		device: ctx.Device,
		cache:  cache,
	}
}

func (p *PipelineCache[C]) Get(conf C) (CachedPipeline, error) {
	cached, ok := p.cache.Get(conf)
	if ok {
		return cached, nil
	}

	pipeline, err := conf.Specialize(p.device)
	if err != nil {
		return CachedPipeline{}, fmt.Errorf("build pipeline: %w", err)
	}

	bindGroupsCache, _ := lru.NewWithEvict[uint32, *wgpu.BindGroupLayout](16, releaseBindGroupLayoutOnEviction)

	pc := CachedPipeline{Pipeline: pipeline, bindGroups: bindGroupsCache}
	p.cache.Add(conf, pc)

	return pc, nil
}

func releasePipelineOnEviction[C any](_config C, pipe CachedPipeline) {
	pipe.bindGroups.Purge()
	pipe.Pipeline.Release()
}

func releaseBindGroupLayoutOnEviction(_ uint32, ev *wgpu.BindGroupLayout) {
	ev.Release()
}
