package orion

import (
	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/pulse"
)

var currentWindow global[glimpse.Window]
var currentContext global[*pulse.Context]
var currentView global[*pulse.View]

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

func (g *global[T]) Get() T {
	if !g.hasValue {
		panic("must only be called after RunGame")
	}

	return g.value
}

func (g *global[T]) reset() {
	var tZero T
	g.value = tZero
	g.hasValue = false
}
