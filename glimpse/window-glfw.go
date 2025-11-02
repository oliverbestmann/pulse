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
	win  *glfw.Window
	prof interface{ Stop() }
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

	win := &glfwWindow{
		win:  window,
		prof: profile.Start(profile.MemProfile),
	}

	return win, nil
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

func (g *glfwWindow) Run(render func()) {
	for !g.win.ShouldClose() {
		glfw.PollEvents()
		render()
	}
}
