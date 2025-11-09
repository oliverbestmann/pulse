package pulse

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
	lru "github.com/hashicorp/golang-lru/v2"
)

var samplerCache, _ = lru.NewWithEvict[wgpu.SamplerDescriptor, *wgpu.Sampler](16, samplerCacheOnEvict)

func samplerCacheOnEvict(key wgpu.SamplerDescriptor, value *wgpu.Sampler) {
	value.Release()
}

// CachedSampler returns a sampler matching your description. The sampler may be cached,
// you  must not call wgpu.Sampler.Release() on it.
func CachedSampler(dev *wgpu.Device, desc wgpu.SamplerDescriptor) (*wgpu.Sampler, error) {
	cachedSampler, ok := samplerCache.Get(desc)
	if ok {
		return cachedSampler, nil
	}

	// create a new device
	sampler, err := dev.CreateSampler(&desc)
	if err != nil {
		return nil, fmt.Errorf("create sampler: %w", err)
	}

	samplerCache.Add(desc, sampler)

	return sampler, nil
}
