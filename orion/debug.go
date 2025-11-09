package orion

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/pulse"
)

type frame struct {
	Total time.Duration

	GetCurrentTexture time.Duration
	GameUpdate        time.Duration
	GameDraw          time.Duration
}

var DebugOverlay debugOverlay

type debugOverlay struct {
	frameCount int
	frames     [60 * 10]frame

	white *Image

	timeStartFrame             time.Time
	timeStartGameDraw          time.Time
	timeStartGameUpdate        time.Time
	timeStartGetCurrentTexture time.Time
	timeEndFrame               time.Time

	mem runtime.MemStats
}

func (d *debugOverlay) StartFrame() {
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
	d.timeStartGetCurrentTexture = time.Now()
}

func (d *debugOverlay) StartGameUpdate() {
	d.timeStartGameUpdate = time.Now()
}

func (d *debugOverlay) StartGameDraw() {
	d.timeStartGameDraw = time.Now()
}

func (d *debugOverlay) EndFrame() {
	d.timeEndFrame = time.Now()

	runtime.ReadMemStats(&d.mem)
}

func (d *debugOverlay) Draw(target *Image) {
	text := d.buildText()

	if d.white == nil {
		d.white = NewImage(1, 1, nil)
		d.white.Clear(pulse.ColorWhite)
	}

	DebugText(target, text, &DebugTextOptions{
		Transform:  glm.TranslationMat3[float32](16, 16),
		ColorScale: ColorScale{},
	})

	binWidth := float32(target.Width()) / float32(len(d.frames))
	timeScale := float32(30) / (1.0 / 60.0)

	for idx, frame := range d.frames {
		x := float32(idx) * binWidth
		y := float32(target.Height())

		rect := func(duration time.Duration, colorScale ColorScale) {
			height := float32(duration.Seconds()) * timeScale

			target.DrawImage(d.white, &DrawImageOptions{
				ColorScale: colorScale,
				Transform:  glm.TranslationMat3(x, y).Scale(binWidth, -height),
			})

			y -= height
		}

		rect(frame.GameUpdate, ColorScaleRGBA(0.25, 0.25, 1.0, 0.85))
		rect(frame.GameDraw, ColorScaleRGBA(0.25, 1.0, 0.25, 0.85))
		rect(frame.GetCurrentTexture, ColorScaleRGBA(0.5, 0.5, 0.5, 0.5))

		remaining := frame.Total - frame.GameUpdate - frame.GameDraw - frame.GetCurrentTexture
		rect(remaining, ColorScaleRGBA(0.25, 0.25, 0.25, 0.5))
	}

	if fps := float32(d.fps()); fps > 0 {
		y := float32(target.Height()) - timeScale/fps

		target.DrawImage(d.white, &DrawImageOptions{
			Transform: glm.TranslationMat3(0, y).Scale(float32(target.Width()), 1),
		})
	}
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
	}

	return strings.Join(lines, "\n")
}
