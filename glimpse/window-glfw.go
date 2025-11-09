//go:build !js

package glimpse

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/cogentcore/webgpu/wgpuglfw"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pkg/profile"
)

type glfwWindow struct {
	win   *glfw.Window
	prof  interface{ Stop() }
	input InputState
}

func NewWindow(width, height int, title string) (Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("initialize glfw: %w", err)
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create window: %w", err)
	}

	w := &glfwWindow{
		win:  window,
		prof: profile.Start(profile.MemProfile),
	}

	window.SetKeyCallback(func(_win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		keyCode := KeyCode(key)

		if action == glfw.Repeat {
			return
		}

		switch action {
		case glfw.Press:
			w.input.Keys.press(keyCode)

		case glfw.Release:
			w.input.Keys.release(keyCode)
		}
	})

	window.SetMouseButtonCallback(func(_win *glfw.Window, btn glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		button := MouseButton(btn)

		switch action {
		case glfw.Press:
			w.input.Mouse.press(button)
		case glfw.Release:
			w.input.Mouse.release(button)
		}
	})

	window.SetCursorPosCallback(func(_win *glfw.Window, xpos float64, ypos float64) {
		w.input.Mouse.position(float32(xpos), float32(ypos))
	})

	return w, nil
}

func (g *glfwWindow) ShouldClose() bool {
	return g.win.ShouldClose()
}

func (g *glfwWindow) GetSize() (uint32, uint32) {
	width, height := g.win.GetSize()
	return uint32(width), uint32(height)
}

func (g *glfwWindow) SurfaceDescriptor() *wgpu.SurfaceDescriptor {
	return wgpuglfw.GetSurfaceDescriptor(g.win)
}

func (g *glfwWindow) Terminate() {
	g.prof.Stop()
	g.win.Destroy()
	glfw.Terminate()
}

func (g *glfwWindow) Run(render func(input UpdateInputState) error) error {
	var updateInputState UpdateInputState = func() InputState {
		g.input.nextTick()
		glfw.PollEvents()
		return g.input
	}

	for !g.win.ShouldClose() {
		if err := render(updateInputState); err != nil {
			return err
		}
	}

	return nil
}
