package orion

import (
	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/glm"
)

type KeyCode = glimpse.Key
type MouseButton = glimpse.MouseButton

func MousePositionRaw() glm.Vec2f {
	inputState := currentInputState.Get()

	return glm.Vec2f{
		inputState.Mouse.CursorX,
		inputState.Mouse.CursorY,
	}
}

func MousePosition() glm.Vec2f {
	raw := MousePositionRaw()

	return currentScreenTransformInv.Get().
		Transform(raw.Extend(1)).
		Truncate()
}

func IsKeyPressed(key KeyCode) bool {
	inputState := currentInputState.Get()
	return inputState.Keys.Pressed[key]
}

func IsKeyJustPressed(key KeyCode) bool {
	inputState := currentInputState.Get()
	return inputState.Keys.JustPressed[key]
}

func IsKeyJustReleased(key KeyCode) bool {
	inputState := currentInputState.Get()
	return inputState.Keys.JustReleased[key]
}

func IsMouseButtonPressed(button MouseButton) bool {
	inputState := currentInputState.Get()
	return inputState.Mouse.Pressed[button]
}

func IsMouseButtonJustPressed(button MouseButton) bool {
	inputState := currentInputState.Get()
	return inputState.Mouse.JustPressed[button]
}

func IsMouseButtonJustReleased(button MouseButton) bool {
	inputState := currentInputState.Get()
	return inputState.Mouse.JustReleased[button]
}
