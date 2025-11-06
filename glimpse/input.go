package glimpse

import "log/slog"

type UpdateInputState func() InputState

type KeyCode uint32

type MouseButton uint32

type KeysState struct {
	// the keys that are currently marked as "pressed"
	Pressed map[KeyCode]bool

	// keys that where just pressed after the last call to nextTick()
	JustPressed map[KeyCode]bool

	// keys that were just released after the last call to nextTick()
	JustReleased map[KeyCode]bool
}

func (k *KeysState) press(keyCode KeyCode) {
	slog.Info("Key just pressed", slog.Int("keyCode", int(keyCode)))

	setTrue(&k.Pressed, keyCode)
	setTrue(&k.JustPressed, keyCode)
}

func (k *KeysState) release(keyCode KeyCode) {
	setFalse(&k.Pressed, keyCode)
	setTrue(&k.JustReleased, keyCode)
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
