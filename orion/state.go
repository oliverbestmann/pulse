package orion

import (
	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/vyn"
	"github.com/oliverbestmann/pulse/wx"
)

var currentWindow global[vyn.Window]
var currentContext global[*wx.Context]
var currentView global[*wx.Surface]
var currentInputState global[vyn.InputState]

var currentScreenTransform global[glm.Mat3f]
var currentScreenTransformInv global[glm.Mat3f]

type global[T any] struct {
	value    T
	hasValue bool
}

func (g *global[T]) set(value T) *global[T] {
	if g.hasValue {
		panic("value already set")
	}

	g.value = value
	g.hasValue = true
	return g
}

func (g *global[T]) reset() {
	var tZero T
	g.value = tZero
	g.hasValue = false
}

func (g *global[T]) Get() T {
	if !g.hasValue {
		panic("must only be called after RunGame")
	}

	return g.value
}

// CurrentContext exposes the current webgpu context. This can be used
// to build your own pipelines and render passes.
func CurrentContext() *wx.Context {
	return currentContext.Get()
}
