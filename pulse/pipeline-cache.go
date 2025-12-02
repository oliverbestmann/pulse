package pulse

import (
	"github.com/hashicorp/golang-lru/v2"
	"github.com/oliverbestmann/webgpu/wgpu"
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
	Specialize(def *wgpu.Device) *wgpu.RenderPipeline
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

func (p *PipelineCache[C]) Get(conf C) CachedPipeline {
	cached, ok := p.cache.Get(conf)
	if ok {
		return cached
	}

	pipeline := conf.Specialize(p.device)

	bindGroupsCache, _ := lru.NewWithEvict[uint32, *wgpu.BindGroupLayout](16, releaseBindGroupLayoutOnEviction)

	pc := CachedPipeline{Pipeline: pipeline, bindGroups: bindGroupsCache}
	p.cache.Add(conf, pc)

	return pc
}

func releasePipelineOnEviction[C any](_config C, pipe CachedPipeline) {
	pipe.bindGroups.Purge()
	pipe.Pipeline.Release()
}

func releaseBindGroupLayoutOnEviction(_ uint32, ev *wgpu.BindGroupLayout) {
	ev.Release()
}
