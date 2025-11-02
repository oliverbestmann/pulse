package pulse

import (
	_ "embed"
	"fmt"

	_ "image/png"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/oliverbestmann/go3d/glm"
)

//go:embed font.png
var _font_png []byte

type TextCommands struct {
	texture *Texture
	sprites *SpriteCommands
}

func NewTextCommands(ctx *Context, sprites *SpriteCommands) (*TextCommands, error) {
	texture, err := DecodeTextureFromMemory(ctx, _font_png)
	if err != nil {
		return nil, fmt.Errorf("load font texture: %w", err)
	}

	return &TextCommands{texture, sprites}, nil
}

func (t *TextCommands) DrawText(dest *RenderTarget, text string, posX, posY float32) error {
	scale := glm.ScaleMat3[float32](6, 10)

	opts := DrawImageOptions{
		Color:        ColorWhite,
		FilterMode:   wgpu.FilterModeNearest,
		BlendState:   wgpu.BlendStateAlphaBlending,
		AddressModeU: wgpu.AddressModeClampToEdge,
		AddressModeV: wgpu.AddressModeClampToEdge,
	}

	baseX := posX

	for _, ch := range text {
		switch {
		case ch == ' ':
			posX += 6
			continue

		case ch == '\t':
			const tabWidth = 6 * 8
			posX = float32(int(posX+6*4) / tabWidth * tabWidth)
			continue

		case ch == '\n':
			posX = baseX
			posY += 16
			continue

		case ch < 32:
			continue
		}

		posCh, ok := chars[ch]
		if !ok {
			// substitute with question mark
			posCh = chars['?']
		}

		charTexture := t.texture.SubTexture(posCh, glm.Vec2[uint32]{6, 10})

		translation := glm.TranslationMat3(posX, posY)

		// draw shadow
		opts.Color = Color{0, 0, 0, 1}
		opts.Transform = translation.Translate(1, 1).Mul(scale)
		if err := t.sprites.DrawImage(dest, charTexture, opts); err != nil {
			return fmt.Errorf("draw character %q: %w", ch, err)
		}

		// draw the actual text
		opts.Color = Color{1, 1, 1, 1}
		opts.Transform = translation.Mul(scale)
		if err := t.sprites.DrawImage(dest, charTexture, opts); err != nil {
			return fmt.Errorf("draw character %q: %w", ch, err)
		}

		// advance position by one char
		posX += 6
	}

	return nil
}

var chars = map[rune]glm.Vec2[uint32]{
	65:  {0, 0},
	66:  {6, 0},
	67:  {12, 0},
	68:  {18, 0},
	69:  {24, 0},
	70:  {30, 0},
	71:  {36, 0},
	72:  {42, 0},
	73:  {48, 0},
	74:  {54, 0},
	75:  {60, 0},
	76:  {66, 0},
	77:  {72, 0},
	78:  {0, 10},
	79:  {6, 10},
	80:  {12, 10},
	81:  {18, 10},
	82:  {24, 10},
	83:  {30, 10},
	84:  {36, 10},
	85:  {42, 10},
	86:  {48, 10},
	87:  {54, 10},
	88:  {60, 10},
	89:  {66, 10},
	90:  {72, 10},
	97:  {0, 20},
	98:  {6, 20},
	99:  {12, 20},
	100: {18, 20},
	101: {24, 20},
	102: {30, 20},
	103: {36, 20},
	104: {42, 20},
	105: {48, 20},
	106: {54, 20},
	107: {60, 20},
	108: {66, 20},
	109: {72, 20},
	110: {0, 30},
	111: {6, 30},
	112: {12, 30},
	113: {18, 30},
	114: {24, 30},
	115: {30, 30},
	116: {36, 30},
	117: {42, 30},
	118: {48, 30},
	119: {54, 30},
	120: {60, 30},
	121: {66, 30},
	122: {72, 30},
	48:  {0, 40},
	49:  {6, 40},
	50:  {12, 40},
	51:  {18, 40},
	52:  {24, 40},
	53:  {30, 40},
	54:  {36, 40},
	55:  {42, 40},
	56:  {48, 40},
	57:  {54, 40},
	43:  {60, 40},
	45:  {66, 40},
	61:  {72, 40},
	40:  {0, 50},
	41:  {6, 50},
	91:  {12, 50},
	93:  {18, 50},
	123: {24, 50},
	125: {30, 50},
	60:  {36, 50},
	62:  {42, 50},
	47:  {48, 50},
	42:  {54, 50},
	58:  {60, 50},
	35:  {66, 50},
	37:  {72, 50},
	33:  {0, 60},
	63:  {6, 60},
	46:  {12, 60},
	44:  {18, 60},
	39:  {24, 60},
	34:  {30, 60},
	64:  {36, 60},
	38:  {42, 60},
	36:  {48, 60},
	32:  {54, 60},
}
