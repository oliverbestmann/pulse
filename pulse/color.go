package pulse

import "github.com/oliverbestmann/pulse/glm"

// Color is an RGBA color value.
type Color = glm.Vec4f

var ColorWhite = Color{1, 1, 1, 1}
var ColorBlack = Color{0, 0, 0, 1}
var ColorTransparent = Color{0, 0, 0, 0}
