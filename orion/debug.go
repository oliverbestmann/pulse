package orion

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/oliverbestmann/pulse/glm"
	"github.com/oliverbestmann/pulse/pulse"
)

type frame struct {
	Total time.Duration

	GetCurrentTexture time.Duration
	GameUpdate        time.Duration
	GameDraw          time.Duration
}

var DebugOverlay debugOverlay

type debugOverlay struct {
	RunGC   bool
	enabled bool

	frameCount int
	frames     [60 * 10]frame

	white *Image

	timeStartFrame             time.Time
	timeStartGameDraw          time.Time
	timeStartGameUpdate        time.Time
	timeStartGetCurrentTexture time.Time
	timeEndFrame               time.Time

	mem runtime.MemStats

	// previous number of cycles
	memNumGC            uint32
	objectsAliveAfterGC uint64
}

func (d *debugOverlay) Enable(enable bool) {
	d.enabled = enable
	d.timeStartFrame = time.Time{}
}

func (d *debugOverlay) Toggle() {
	d.Enable(!d.enabled)
}

func (d *debugOverlay) StartFrame() {
	if !d.enabled {
		return
	}

	now := time.Now()

	if !d.timeStartFrame.IsZero() {
		d.frames[d.frameCount%len(d.frames)] = frame{
			Total:             now.Sub(d.timeStartFrame),
			GetCurrentTexture: d.timeStartGameUpdate.Sub(d.timeStartFrame),
			GameUpdate:        d.timeStartGameDraw.Sub(d.timeStartGameUpdate),
			GameDraw:          d.timeEndFrame.Sub(d.timeStartGameDraw),
		}

		d.frameCount += 1
	}

	d.timeStartFrame = now
}

func (d *debugOverlay) StartGetCurrentTexture() {
	if !d.enabled {
		return
	}

	d.timeStartGetCurrentTexture = time.Now()
}

func (d *debugOverlay) StartGameUpdate() {
	if !d.enabled {
		return
	}

	d.timeStartGameUpdate = time.Now()
}

func (d *debugOverlay) StartGameDraw() {
	if !d.enabled {
		return
	}

	d.timeStartGameDraw = time.Now()
}

func (d *debugOverlay) EndFrame() {
	if !d.enabled {
		return
	}

	d.timeEndFrame = time.Now()

	if d.RunGC {
		if d.frameCount%10 == 0 {
			runtime.GC()
		}
	}

	runtime.ReadMemStats(&d.mem)

	if d.mem.NumGC > d.memNumGC {
		d.memNumGC = d.mem.NumGC
		d.objectsAliveAfterGC = d.mem.HeapObjects
	}
}

func (d *debugOverlay) Draw(target *Image) {
	if !d.enabled {
		return
	}

	if d.white == nil {
		d.white = NewImage(1, 1, nil)
		d.white.Clear(pulse.ColorWhite)
	}

	DebugText(target, d.buildText(), &DebugTextOptions{
		Transform:  glm.TranslationMat3[float32](16, 16),
		ColorScale: ColorScale{},
	})

	d.drawFrameStats(target)
}

func (d *debugOverlay) drawFrameStats(target *Image) {
	binWidth := float32(target.Width()) / float32(len(d.frames))
	timeScale := float32(30) / (1.0 / 60.0)

	vertices := make([]Vertex2d, 0, 6*4*len(d.frames)+6)

	for idx, frame := range d.frames {
		x := float32(idx) * binWidth
		y := float32(target.Height())

		rect := func(duration time.Duration, colorScale ColorScale) {
			height := float32(duration.Seconds()) * timeScale

			vertices = append(vertices,
				Vertex2d{
					Position: glm.Vec2f{x, y},
					Color:    colorScale,
				},
				Vertex2d{
					Position: glm.Vec2f{x, y - height},
					Color:    colorScale,
				},
				Vertex2d{
					Position: glm.Vec2f{x + binWidth, y - height},
					Color:    colorScale,
				},
				Vertex2d{
					Position: glm.Vec2f{x, y},
					Color:    colorScale,
				},
				Vertex2d{
					Position: glm.Vec2f{x + binWidth, y - height},
					Color:    colorScale,
				},
				Vertex2d{
					Position: glm.Vec2f{x + binWidth, y},
					Color:    colorScale,
				},
			)

			y -= height
		}

		rect(frame.GameUpdate, pulse.ColorLinearRGBA(0.25, 0.25, 1.0, 0.85))
		rect(frame.GameDraw, pulse.ColorLinearRGBA(0.25, 1.0, 0.25, 0.85))
		rect(frame.GetCurrentTexture, pulse.ColorLinearRGBA(0.5, 0.5, 0.5, 0.5))

		remaining := frame.Total - frame.GameUpdate - frame.GameDraw - frame.GetCurrentTexture
		rect(remaining, pulse.ColorLinearRGBA(0.25, 0.25, 0.25, 0.5))
	}

	if fps := float32(d.fps()); fps > 0 {
		y := float32(target.Height()) - timeScale/fps - 0.5

		vertices = append(vertices,
			Vertex2d{
				Position: glm.Vec2f{0, y},
				Color:    pulse.ColorWhite,
			},
			Vertex2d{
				Position: glm.Vec2f{0, y + 1},
				Color:    pulse.ColorWhite,
			},
			Vertex2d{
				Position: glm.Vec2f{float32(target.Width()), y + 1},
				Color:    pulse.ColorWhite,
			},
			Vertex2d{
				Position: glm.Vec2f{0, y},
				Color:    pulse.ColorWhite,
			},
			Vertex2d{
				Position: glm.Vec2f{float32(target.Width()), y + 1},
				Color:    pulse.ColorWhite,
			},
			Vertex2d{
				Position: glm.Vec2f{float32(target.Width()), y},
				Color:    pulse.ColorWhite,
			},
		)
	}

	target.DrawTriangles(vertices, nil)
}

func (d *debugOverlay) fps() float64 {
	// calculate the average frame time
	var frameCount int
	var totalTime time.Duration

	for _, frame := range d.frames {
		if frame.Total > 0 {
			frameCount += 1
			totalTime += frame.Total
		}
	}

	if frameCount == 0 {
		return 0
	}

	averageFrameTime := totalTime / time.Duration(frameCount)

	// calculate the frames per second
	return 1.0 / averageFrameTime.Seconds()
}

func (d *debugOverlay) buildText() string {

	lastCycle := (d.mem.NumGC + 255) % 256
	lastCycleDur := time.Duration(d.mem.PauseNs[lastCycle])

	lines := []string{
		fmt.Sprintf("FPS: %1.2f", d.fps()),
		fmt.Sprintf("Frames: %d", d.frameCount),
		fmt.Sprintf(""),
		fmt.Sprintf("Memory"),
		fmt.Sprintf("  Heap Objects: %d", d.mem.HeapObjects),
		fmt.Sprintf("  Heap InUse:   %1.2fmb", float64(d.mem.HeapInuse)/(1024.0*1024.0)),
		fmt.Sprintf("  Stack InUse:  %1.2fmb", float64(d.mem.StackInuse)/(1024.0*1024.0)),
		fmt.Sprintf(""),
		fmt.Sprintf("GC:"),
		fmt.Sprintf("  Cycles:   %d", d.mem.NumGC),
		fmt.Sprintf("  Fraction: %1.2f%%", d.mem.GCCPUFraction*100),
		fmt.Sprintf("  Duration: %1.2fms", lastCycleDur.Seconds()*1000),
		fmt.Sprintf("  Alive:    %d", d.objectsAliveAfterGC),
	}

	return strings.Join(lines, "\n")
}
