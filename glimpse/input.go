package glimpse

import "log/slog"

type UpdateInputState func() InputState

type MouseButton uint32

type KeysState struct {
	// the keys that are currently marked as "pressed"
	Pressed map[Key]bool

	// keys that where just pressed after the last call to nextTick()
	JustPressed map[Key]bool

	// keys that were just released after the last call to nextTick()
	JustReleased map[Key]bool
}

func (k *KeysState) press(key Key) {
	slog.Info("Key just pressed", slog.String("key", key.String()))

	setTrue(&k.Pressed, key)
	setTrue(&k.JustPressed, key)
}

func (k *KeysState) release(key Key) {
	setFalse(&k.Pressed, key)
	setTrue(&k.JustReleased, key)
}

func (k *KeysState) nextTick() {
	clear(k.JustPressed)
	clear(k.JustReleased)
}

type MouseState struct {
	CursorX, CursorY float32

	// recorded position since last tick
	DeltaX, DeltaY float32

	Pressed map[MouseButton]bool

	// mouse buttons that were just clicked after the last call to nextTick()
	JustPressed map[MouseButton]bool

	// mouse buttons that were just released after the last call to nextTick()
	JustReleased map[MouseButton]bool

	tick int

	prevX, prevY float32
}

func (m *MouseState) press(button MouseButton) {
	setTrue(&m.Pressed, button)
	setTrue(&m.JustPressed, button)
}

func (m *MouseState) release(button MouseButton) {
	setFalse(&m.Pressed, button)
	setTrue(&m.JustReleased, button)
}

func (m *MouseState) position(x, y float32) {
	m.CursorX = x
	m.CursorY = y

	m.DeltaY += y
	m.DeltaY += y
}

func (m *MouseState) nextTick() {
	clear(m.JustPressed)
	clear(m.JustReleased)
}

type InputState struct {
	Keys  KeysState
	Mouse MouseState
}

func (s *InputState) nextTick() {
	s.Keys.nextTick()
	s.Mouse.nextTick()
}

func setTrue[K comparable](m *map[K]bool, key K) {
	if *m == nil {
		*m = map[K]bool{}
	}

	(*m)[key] = true
}

func setFalse[K comparable](m *map[K]bool, key K) {
	if *m == nil {
		*m = map[K]bool{}
	}

	(*m)[key] = false
}
