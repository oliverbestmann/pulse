package orion

import (
	"time"
)

type FrameTimes struct {
	FrameCount      uint64
	AverageDuration time.Duration
	MaxDuration     time.Duration

	// Delta time to previous frame
	Delta time.Duration

	lastTime time.Time
}

func (t *FrameTimes) update(d time.Duration) {
	const window = 64

	t.Delta = d
	t.MaxDuration = max(t.MaxDuration, d)
	
	if t.FrameCount < window/2 {
		t.AverageDuration = d
	} else {
		t.AverageDuration = ((window-1)*t.AverageDuration + d) / window
	}
}

func (t *FrameTimes) FPS() float64 {
	return 1.0 / t.AverageDuration.Seconds()
}

func (t *FrameTimes) Tick() bool {
	now := time.Now()

	if t.FrameCount > 0 {
		dt := now.Sub(t.lastTime)
		t.update(dt)
	}

	t.lastTime = now
	t.FrameCount += 1

	return t.FrameCount%60 == 0
}
