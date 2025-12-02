package pulse

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/oliverbestmann/webgpu/wgpu"
)

var samplerCache, _ = lru.NewWithEvict[wgpu.SamplerDescriptor, *wgpu.Sampler](16, samplerCacheOnEvict)

func samplerCacheOnEvict(key wgpu.SamplerDescriptor, value *wgpu.Sampler) {
	value.Release()
}

// CachedSampler returns a sampler matching your description. The sampler may be cached,
// you  must not call wgpu.Sampler.Release() on it.
func CachedSampler(dev *wgpu.Device, desc wgpu.SamplerDescriptor) *wgpu.Sampler {
	cachedSampler, ok := samplerCache.Get(desc)
	if ok {
		return cachedSampler
	}

	// create a new sampler
	sampler := dev.CreateSampler(&desc)

	// and cache it for the next access
	samplerCache.Add(desc, sampler)

	return sampler
}
