//go:build js

package glimpse

import (
	"syscall/js"

	"github.com/cogentcore/webgpu/wgpu"
)

type jsWindow struct {
	canvas js.Value
}

func NewWindow(width, height int, title string) (Window, error) {
	document := js.Global().Get("document")
	canvas := document.Call("createElement", "canvas")
	document.Get("body").Call("appendChild", canvas)

	document.Set("title", title)

	canvas.Set("style", "width:100vw; height:100vh")

	win := &jsWindow{
		canvas: canvas,
	}

	return win, nil
}

func (g *jsWindow) ShouldClose() bool {
	return false
}

func (g *jsWindow) GetSize() (uint32, uint32) {
	ratio := js.Global().Get("devicePixelRatio").Float()

	vv := js.Global().Get("visualViewport")
	width := vv.Get("width").Int()
	height := vv.Get("height").Int()
	return uint32(float64(width) * ratio), uint32(float64(height) * ratio)
}

func (g *jsWindow) SurfaceDescriptor() *wgpu.SurfaceDescriptor {
	return &wgpu.SurfaceDescriptor{Canvas: g.canvas}
}

func (g *jsWindow) Terminate() {
	// do nothing
}

func (g *jsWindow) Run(render func() error) error {
	errCh := make(chan error, 1)

	renderOnce := func() bool {
		resizeCanvas(g.canvas)

		if err := render(); err != nil {
			errCh <- err
			return false
		}

		return true
	}

	var renderAndSchedule js.Func

	renderAndSchedule = js.FuncOf(func(this js.Value, args []js.Value) any {
		// we must not block in a FuncOf callback. spawn a go routine and call
		// requestAnimationFrame later from there
		go func() {
			if renderOnce() {
				js.Global().Call("requestAnimationFrame", renderAndSchedule)
			}
		}()

		return nil
	})

	defer renderAndSchedule.Release()

	// trigger the async render loop
	renderAndSchedule.Invoke()

	// block until we get an error
	return <-errCh
}

func resizeCanvas(canvas js.Value) {
	vv := js.Global().Get("visualViewport")
	viewWidth := vv.Get("width").Float()
	viewHeight := vv.Get("height").Float()

	ratio := js.Global().Get("devicePixelRatio").Float()

	canvas.Set("width", viewWidth*ratio)
	canvas.Set("height", viewHeight*ratio)
}
