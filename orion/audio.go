package orion

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/oto/v3"
)

type Sample = float32
type StereoSample [2]Sample

const StereoSamplesPerSecond = 48000

const stereoSampleSize = unsafe.Sizeof(StereoSample{})

var context = sync.OnceValue(func() *oto.Context {
	context, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   StereoSamplesPerSecond,
		Format:       oto.FormatFloat32LE,
		ChannelCount: 2,
		BufferSize:   32 * time.Millisecond,
	})

	if err != nil {
		slog.Warn("Failed to initialize AudioContext", slog.String("error", err.Error()))
		return nil
	}

	go func() {
		<-ready
		slog.Info("AudioContext is ready")
	}()

	go playersCleanup()

	return context
})

func playersCleanup() {
	runOnce := func() {
		playersMu.Lock()
		defer playersMu.Unlock()

		now := time.Now()

		for player, checkWhen := range players {
			if checkWhen.Before(now) {
				continue
			}

			if !player.IsPlaying() {
				delete(players, player)
			}
		}
	}

	for {
		time.Sleep(1 * time.Second)
		runOnce()
	}
}

var playersMu sync.Mutex
var players = map[*oto.Player]time.Time{}

type AudioPlayer interface {
	Pause()
	Play()
	IsPlaying() bool
	Volume() float32
	SetVolume(volume float32)
	TrySeekTo(stereoSampleIdx int) error
}

func PlayAudio(samples []StereoSample) AudioPlayer {
	ctx := context()
	if ctx == nil {
		return &noopPlayer{}
	}

	// cast the sample slice to a byte array that the reader can read from
	ptr := (*byte)(unsafe.Pointer(unsafe.SliceData(samples)))
	buf := unsafe.Slice(ptr, uintptr(len(samples))*stereoSampleSize)

	p := ctx.NewPlayer(bytes.NewReader(buf))
	p.Play()

	playersMu.Lock()
	players[p] = time.Now()
	playersMu.Unlock()

	return &player{player: p}
}

type SampleReader interface {
	Read(samples []StereoSample) (n int64, err error)
}

func StreamAudio(r SampleReader) AudioPlayer {
	ctx := context()
	if ctx == nil {
		return &noopPlayer{}
	}

	p := ctx.NewPlayer(&readToBytes{sr: r})
	p.Play()

	playersMu.Lock()
	players[p] = time.Now()
	playersMu.Unlock()

	return &player{player: p}
}

func SuspendAudio() {
	ctx := context()
	if ctx != nil {
		_ = ctx.Suspend()
	}
}

func ResumeAudio() {
	ctx := context()
	if ctx != nil {
		_ = ctx.Resume()
	}
}

type readToBytes struct {
	sr SampleReader
}

func (r *readToBytes) Read(p []byte) (n int, err error) {
	if len(p)%int(stereoSampleSize) != 0 {
		return 0, errors.New("can only read multiple of sample size")
	}

	ptr := (*StereoSample)(unsafe.Pointer(unsafe.SliceData(p)))
	buf := unsafe.Slice(ptr, uintptr(len(p))/stereoSampleSize)

	nSamples, err := r.sr.Read(buf)
	n = int(nSamples) * int(stereoSampleSize)
	return n, err
}

type player struct {
	player *oto.Player
}

func (p *player) Pause() {
	playersMu.Lock()
	delete(players, p.player)
	playersMu.Unlock()

	p.player.Pause()
}

func (p *player) Play() {
	p.player.Play()

	playersMu.Lock()
	players[p.player] = time.Now()
	playersMu.Unlock()
}

func (p *player) IsPlaying() bool {
	return p.player.IsPlaying()
}

func (p *player) Volume() float32 {
	return float32(p.player.Volume())
}

func (p *player) SetVolume(volume float32) {
	p.player.SetVolume(float64(volume))
}

func (p *player) TrySeekTo(stereoSampleIdx int) error {
	_, err := p.player.Seek(int64(stereoSampleIdx)*int64(stereoSampleSize), io.SeekStart)
	return err
}

type noopPlayer struct{}

func (n *noopPlayer) Pause() {
}

func (n *noopPlayer) Play() {
}

func (n *noopPlayer) IsPlaying() bool {
	return false
}

func (n *noopPlayer) Volume() float32 {
	return 1.0
}

func (n *noopPlayer) SetVolume(volume float32) {
}

func (n *noopPlayer) TrySeekTo(stereoSampleIdx int) error {
	return nil
}
