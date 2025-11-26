package main

import (
	_ "embed"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"time"
	"unsafe"

	_ "image/png"

	"github.com/furui/fastnoiselite-go"
	b2 "github.com/oliverbestmann/box2d-go"
	"github.com/oliverbestmann/go3d/glimpse"
	"github.com/oliverbestmann/go3d/glm"
	"github.com/oliverbestmann/go3d/orion"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed ship.png
var _ship []byte

//go:embed player.png
var _player []byte

var ColorBlack = orion.Color{0.1, 0.1, 0.2, 1.0}
var ColorWhite = orion.Color{0.95, 0.90, 0.8, 1.0}
var ColorAccent = orion.Color{0.7, 0.16, 0.35, 1.0}

type Asteroid struct {
	Body     b2.Body
	Vertices []orion.Vertex2d
}

type Player struct {
	Body     b2.Body
	Shape    b2.Shape
	Position glm.Vec2f
	Radius   float32
	Rotation glm.Rad

	FlipX     float32
	LightSize float32

	LightOn   bool
	MarkerOn  bool
	Thrusting bool
}

type Ping struct {
	Origin glm.Vec2f
	Color  orion.Color
	Size   float32
	Width  float32
	Speed  float32
}

type Particle struct {
	Position glm.Vec2f
	Velocity glm.Vec2f
	Value    float32
}

type Game struct {
	world b2.World

	debug bool

	lastTime  time.Time
	asteroids []Asteroid
	player    Player
	beacon    glm.Vec2f
	toScreen  glm.Mat3f
	toWorld   glm.Mat3f

	pingVertices []orion.Vertex2d

	playerVertices []orion.Vertex2d
	beaconVertices []orion.Vertex2d
	lightVertices  []orion.Vertex2d

	pings           []Ping
	screenWidth     uint32
	screenHeight    uint32
	remainingOxygen float32
	elapsedTime     float32

	particles   []Particle
	rng         *rand.Rand
	timeStepAcc time.Duration

	// true if the first and second automated pingByPlayer were sent out
	pingOne   bool
	pingTwo   bool
	shipImage *orion.Image

	nextBeaconPing float32
	dead           bool
	cameraShake    float32

	noise        *fastnoiselite.FastNoiseLite
	playerImages []*orion.Image
	hitShip      bool

	plNoise orion.AudioPlayer
}

func (g *Game) DrawToSurface(surface, offscreen *orion.Image) {
	orion.DefaultDrawToSurface(surface, offscreen, wgpu.FilterModeNearest)
}

type noise struct {
	offset int
}

func (no *noise) Read(samples []orion.StereoSample) (n int64, err error) {
	for idx := range samples {
		now := float64(no.offset) / orion.StereoSamplesPerSecond

		t1 := float64(no.offset) / orion.StereoSamplesPerSecond * 440
		t2 := float64(no.offset) / orion.StereoSamplesPerSecond * (440 + now)

		sample := float32(math.Sin(t1) + math.Sin(t2)*0.5)

		samples[idx] = orion.StereoSample{
			sample,
			sample,
		}

		no.offset += 1
	}

	return int64(len(samples)), nil
}

func (g *Game) Initialize() error {
	g.lastTime = time.Now()

	g.shipImage, _ = orion.DecodeImageFromBytes(_ship)

	playerImage, _ := orion.DecodeImageFromBytes(_player)

	g.playerImages = []*orion.Image{
		playerImage.SubImage(0, 0, 64, 64),
		playerImage.SubImage(64, 0, 64, 64),
	}

	// we force two pings, each taking 1s of oxygen, reducing the
	// real play time to 20s
	g.remainingOxygen = 22

	g.lightVertices = circleVertices(128)
	g.playerVertices = circleVertices(16)
	g.beaconVertices = circleVertices(16)

	g.player = g.buildPlayer()

	g.beacon = glm.Vec2f{500, -400}

	rng := rand.New(rand.NewPCG(0, 3))
	g.rng = rng

	g.noise = fastnoiselite.NewNoise()
	g.noise.SetNoiseType(fastnoiselite.NoiseTypeOpenSimplex2)
	g.noise.FractalType = fastnoiselite.FractalTypeFBm
	g.noise.Frequency = 5.0
	g.noise.SetFractalOctaves(3)

	for range 64 {
		g.randomAsteroid(rng, 48, 5)
	}

	def := b2.DefaultBodyDef()
	def.Type1 = b2.StaticBody
	def.Position = b2Vec(g.beacon)
	body := g.world.CreateBody(def)
	shape := b2.DefaultShapeDef()
	shape.IsSensor = 1
	shape.EnableSensorEvents = 1
	body.CreateCircleShape(shape, b2.Circle{
		Radius: 32.0,
	})

	g.plNoise = orion.StreamAudio(&noise{})

	return nil
}

func (g *Game) buildPlayer() Player {
	worldDef := b2.DefaultWorldDef()
	worldDef.Gravity = b2.ZeroVec2
	g.world = b2.CreateWorld(worldDef)

	bPlayerDef := b2.DefaultBodyDef()
	bPlayerDef.Type1 = b2.DynamicBody
	bPlayerDef.LinearVelocity = b2.Vec2{X: -32, Y: -16}
	bPlayerDef.LinearDamping = 0.0
	bPlayerDef.Position = b2.Vec2{X: -500, Y: 350}
	body := g.world.CreateBody(bPlayerDef)

	def := b2.DefaultShapeDef()
	def.Density = 1
	def.EnableContactEvents = 1
	def.EnableSensorEvents = 1
	shape := body.CreateCircleShape(def, b2.Circle{Radius: 8.0})

	player := Player{
		Body:      body,
		Shape:     shape,
		Radius:    8.0,
		Position:  toVec(body.GetPosition()),
		LightSize: 100,

		LightOn:  true,
		MarkerOn: true,
	}
	return player
}

func (g *Game) randomAsteroid(rng *rand.Rand, radius, density float32) {
	var position glm.Vec2f

outer:
	for {
		position = randVec(rng).Scale(1024.0)
		for _, a := range g.asteroids {
			if toVec(a.Body.GetPosition()).Sub(position).Length() < 2*radius {
				continue outer
			}
			if g.player.Position.Sub(position).Length() < 3*radius {
				continue outer
			}
			if g.beacon.Sub(position).Length() < 3*radius {
				continue outer
			}
		}

		break
	}

	var vertices []orion.Vertex2d

	var points []glm.Vec2f

	// overall radius of this asteroid
	radius = randf(rng, 0.8, 1.0) * radius

	var steps = rng.IntN(3) + 5
	for idx := range steps {
		scale := 2 * math.Pi / float32(steps)
		angle := glm.Rad(float32(idx) * scale)
		dist := randf(rng, 0.8, 1) * radius

		pos := glm.RotationMat3[float32](angle).Transform2(glm.Vec2f{dist, 0.0})
		points = append(points, pos)
	}

	for idx := 2; idx < len(points); idx++ {
		vertices = append(vertices,
			orion.Vertex2d{Position: points[0]},
			orion.Vertex2d{Position: points[idx-1]},
			orion.Vertex2d{Position: points[idx]},
		)
	}

	hull, ok := b2.ComputeHull(b2Vecs(points))
	if !ok {
		panic("calculate points")
	}

	poly := b2.MakePolygon(hull, 0)

	def := b2.DefaultBodyDef()
	def.Type1 = b2.DynamicBody
	def.LinearVelocity = b2Vec(randVec(rng).Scale(5.0))
	def.LinearDamping = 0
	def.AngularVelocity = randf(rng, -0.3, 0.3)
	def.AngularDamping = 0
	def.Position = b2Vec(position)

	body := g.world.CreateBody(def)

	shape := b2.DefaultShapeDef()
	shape.Material.Restitution = 0.1
	shape.Density = density
	shape.EnableContactEvents = 0
	body.CreatePolygonShape(shape, poly)

	g.asteroids = append(g.asteroids, Asteroid{
		Body:     body,
		Vertices: vertices,
	})
}

func b2Vecs(vecs []glm.Vec2f) []b2.Vec2 {
	b2Vecs := (*b2.Vec2)(unsafe.Pointer(unsafe.SliceData(vecs)))
	return unsafe.Slice(b2Vecs, len(vecs))
}

func (g *Game) Update() error {
	dt, stepCount := g.fixedTimeStep()

	if orion.IsKeyJustPressed(glimpse.KeyD) {
		orion.DebugOverlay.Enable(true)
	}

	for step := range stepCount {
		firstStep := step == 0
		g.elapsedTime += dt

		g.world.Step(dt, 4)

		pl := &g.player

		if !g.hitShip {
			// update player position
			pl.Position = toVec(pl.Body.GetPosition())

			// reduce oxygen...
			g.remainingOxygen -= dt
		}

		if !g.dead && !g.hitShip && orion.IsMouseButtonPressed(orion.MouseButton(0)) {
			mouse := g.toWorld.Transform2(orion.MousePosition())

			delta := mouse.Sub(g.player.Position)
			amount := delta.Length()

			if amount > 32 {
				thrust := delta.Normalize().Scale(12000)
				g.player.Body.ApplyForceToCenter(b2Vec(thrust), 1)

				velocity := toVec(g.player.Body.GetLinearVelocity())
				g.spawn(g.player.Position, delta.Normalize().Scale(-50).Add(velocity))

				pl.Thrusting = true
			} else if firstStep && orion.IsMouseButtonJustPressed(orion.MouseButton(0)) {
				g.pingByPlayer()
			}
		} else {
			pl.Thrusting = false
		}

		if !g.dead && firstStep && orion.IsKeyJustPressed(glimpse.KeySpace) {
			g.pingByPlayer()
		}

		if !g.dead {
			// mirror rotation
			mouse := g.toWorld.Transform2(orion.MousePosition())
			delta := g.player.Position.Sub(mouse)
			angle := glm.Rad(math.Atan2(float64(delta[1]), float64(delta[0])))

			// flip on x if needed
			var flipX float32 = 1
			if delta[0] < 0 {
				flipX = -1
				angle = angle - math.Pi
			}

			pl.Rotation = angle
			pl.FlipX = flipX
		}

		for idx := range g.pings {
			p := &g.pings[idx]
			p.Size += dt * p.Speed
		}

		particles := g.particles
		g.particles = g.particles[:0]
		for _, p := range particles {
			if p.Value < 1.0 {
				p.Value += dt
				p.Position = p.Position.Add(p.Velocity.Scale(dt))
				g.particles = append(g.particles, p)
			}
		}

		if !g.pingOne && g.remainingOxygen < 21.5 {
			g.pingOne = true
			g.pingByPlayer()
		}

		if !g.pingTwo && g.remainingOxygen < 19.5 {
			g.pingTwo = true
			g.pingByPlayer()
		}

		if g.elapsedTime > 3.0 && g.elapsedTime < 7.0 && g.elapsedTime > g.nextBeaconPing {
			g.nextBeaconPing += g.elapsedTime + 8.0
			g.pingByBeacon()
		}

		if rem := g.remainingOxygen; rem < 0 {
			g.dead = true
			g.plNoise.Pause()

			// flicker off
			g.player.LightOn = rem > -1 && g.sampleNoise(g.elapsedTime, 0) > 0
			g.player.MarkerOn = rem > -1 || (rem > -2 && g.sampleNoise(g.elapsedTime, 0) > 0)

			// slow down the players movement now that he is dead
			g.player.Body.SetLinearDamping(0.5)
		}

		if !g.dead {
			for _, event := range g.world.GetContactEvents().BeginEvents {
				p := g.player.Shape.Id
				if event.ShapeIdA != p && event.ShapeIdB != p {
					// not the player
					continue
				}

				velocity := toVec(g.player.Body.GetLinearVelocity())
				trauma := max(0, velocity.Length()-3) / 10
				g.cameraShake = min(1, trauma*trauma)
			}

			for _, event := range g.world.GetSensorEvents().BeginEvents {
				p := g.player.Shape.Id
				if event.VisitorShapeId != p {
					// not the player
					continue
				}

				g.hitShip = true
				g.plNoise.Pause()
			}
		}

		if g.hitShip {
			// expand light cone, and nudge it to the ship
			pl.LightSize += 200 * dt
			pl.Position = pl.Position.Scale(1 - dt).Add(g.beacon.Scale(dt))
		}

		// reduce camera shake
		g.cameraShake *= 0.95
	}

	return nil
}

func (g *Game) pingByPlayer() {
	// each pingByPlayer deducts 1 second
	g.remainingOxygen -= 1.0

	g.pings = append(g.pings, Ping{
		Origin: g.player.Position,
		Color:  ColorWhite,
		Size:   100,
		Width:  100,
		Speed:  500,
	})
}

func (g *Game) pingByBeacon() {
	g.pings = append(g.pings, Ping{
		Origin: g.beacon,
		Color:  ColorAccent,
		Size:   100,
		Width:  50,
		Speed:  1000,
	})
}

func (g *Game) Draw(screen *orion.Image) {
	screen.Clear(ColorBlack)

	g.updateToScreenTransform()

	pl := &g.player

	toScreen := g.toScreen

	// draw pings
	for _, ping := range g.pings {
		g.drawRing(screen, ping.Origin, ping.Size, ping.Size-ping.Width, ping.Color)
	}

	if pl.LightOn {
		// draw the players light
		size := pl.LightSize

		screen.DrawTriangles(g.lightVertices, &orion.DrawTrianglesOptions{
			ColorScale: orion.ColorScaleOf(ColorWhite),
			Transform:  toScreen.Translate(pl.Position.XY()).Scale(size, size),
		})
	}

	// draw particles
	particleVertices := circleVertices(4)
	for _, p := range g.particles {
		if p.Value > 0.5 {
			continue
		}

		screen.DrawTriangles(particleVertices, &orion.DrawTrianglesOptions{
			ColorScale: orion.ColorScaleOf(ColorBlack),
			Transform:  toScreen.Translate(p.Position.XY()).Scale(4.0, 4.0),
		})
	}

	// draw the asteroids
	for _, a := range g.asteroids {
		pos := toVec(a.Body.GetPosition())
		angle := glm.Rad(a.Body.GetRotation().Angle())

		tr := toScreen.Translate(pos.XY()).Rotate(angle)

		screen.DrawTriangles(a.Vertices, &orion.DrawTrianglesOptions{
			ColorScale: orion.ColorScaleOf(ColorBlack),
			Transform:  tr,
		})
	}

	// draw the ship
	{
		unitScale := glm.TranslationMat3[float32](-0.5, -0.5).Scale(g.shipImage.Sizef().Recip().XY())

		angle := glm.Rad(g.elapsedTime * 0.1)

		screen.DrawImage(g.shipImage, &orion.DrawImageOptions{
			ColorScale: orion.ColorScaleOf(ColorBlack),
			Transform:  toScreen.Translate(g.beacon.XY()).Scale(128, 128).Rotate(angle).Mul(unitScale),
		})
	}

	// draw the target beacon
	{
		screen.DrawTriangles(g.beaconVertices, &orion.DrawTrianglesOptions{
			ColorScale: orion.ColorScaleOf(ColorAccent),
			Transform:  toScreen.Translate(g.beacon.XY()).Scale(16, 16),
		})
	}

	// draw the player
	if pl.MarkerOn && !g.hitShip {
		image := g.playerImages[1]

		if !g.dead && pl.Thrusting {
			image = g.playerImages[0]
		}

		// scale image to unit size and move the anchor to the center
		toUnitSize := glm.TranslationMat3[float32](-0.5, -0.5).Scale(1/64.0, 1/64.0)

		// size of the image
		size := pl.Radius * 3

		screen.DrawImage(image, &orion.DrawImageOptions{
			ColorScale: orion.ColorScaleOf(ColorAccent),
			Transform: toScreen.Translate(pl.Position.XY()).
				Rotate(pl.Rotation).
				Scale(size*pl.FlipX, size).
				Mul(toUnitSize),
		})
	}

	var text string
	var color orion.Color

	switch {
	case g.hitShip:
		text = "You've made\nit to safety!"
		color = ColorBlack
	case g.dead:
		text = "You're dead,\ntap to try again..."
		color = ColorWhite
	default:
		text = fmt.Sprintf("Oxygen\n%1.2fsec", max(0, g.remainingOxygen))
		color = ColorWhite
	}

	pos := pl.Position.Add(glm.Vec2f{-32, 64})

	orion.DebugText(screen, text, &orion.DebugTextOptions{
		ColorScale:  orion.ColorScaleOf(color),
		Transform:   toScreen.Translate(pos.XY()).Scale(2.0, 2.0),
		ShadowColor: orion.Color{0, 0, 0, 0},
	})

	// draw debug overlay if enabled
	orion.DebugOverlay.Draw(screen)
}

func (g *Game) updateToScreenTransform() {
	aspect := float32(16.0 / 9.0)

	// keep the screen at 1024px width
	w := float32(g.screenWidth)

	toScreen := glm.ScaleMat3(w/1024.0, w/1024.0)

	// offset camera to put origin to the center of the screen
	toScreen = toScreen.Translate(512, 512.0/aspect)

	// move camera back halfway to the origin
	toScreen = toScreen.Translate(g.player.Position.Scale(0.5).XY())

	// rotate camera shake
	if g.cameraShake > 1e-3 {
		angle := glm.Rad(g.sampleNoise(g.elapsedTime, 1.0) * 0.1 * g.cameraShake)

		offset := glm.Vec2f{
			g.sampleNoise(g.elapsedTime, 2.0) * 32 * g.cameraShake,
			g.sampleNoise(g.elapsedTime, 3.0) * 32 * g.cameraShake,
		}

		toScreen = toScreen.Translate(offset.XY()).Rotate(angle)
	}

	// center camera around the player
	toScreen = toScreen.Translate(g.player.Position.Scale(-1).XY())

	g.toScreen = toScreen
	g.toWorld = toScreen.Invert()
}

func (g *Game) Layout(surfaceWidth, surfaceHeight uint32) orion.LayoutOptions {
	width := surfaceWidth
	height := surfaceWidth * 9 / 16

	g.screenWidth = width
	g.screenHeight = height

	return orion.LayoutOptions{
		Width:  width,
		Height: height,
		MSAA:   true,
	}
}

func main() {
	err := orion.RunGame(orion.RunGameOptions{
		Game:            &Game{},
		WindowWidth:     1024,
		WindowHeight:    600,
		WindowTitle:     "Gravit!",
		WindowResizable: true,
	})

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}

func randf(rng *rand.Rand, lower, upper float32) float32 {
	return rng.Float32()*(upper-lower) + lower
}

func randVec(rng *rand.Rand) glm.Vec2f {
	for {
		x := rng.Float32()*2 - 1
		y := rng.Float32()*2 - 1

		if x*x+y*y <= 1 {
			return glm.Vec2f{x, y}
		}
	}
}

func toVec(v b2.Vec2) glm.Vec2f {
	return glm.Vec2f{v.X, v.Y}
}

func b2Vec(v glm.Vec2f) b2.Vec2 {
	return b2.Vec2{X: v[0], Y: v[1]}
}

func circleVertices(pointCount int) []orion.Vertex2d {
	pointAt := func(idx int, radius float32) orion.Vertex2d {
		angle := glm.Rad(float32(idx%pointCount) * float32(math.Pi*2) / float32(pointCount))
		pos := glm.RotationMat3[float32](angle).Transform2(glm.Vec2f{radius, 0})
		return orion.Vertex2d{Position: pos}
	}

	var vertices []orion.Vertex2d
	for idx := range pointCount {
		vertices = append(vertices,
			pointAt(idx, 0.5),
			pointAt(idx+1, 0.5),
			orion.Vertex2d{},
		)
	}

	return vertices
}

func (g *Game) drawRing(target *orion.Image, pos glm.Vec2f, outer, inner float32, color orion.Color) {
	target.DrawTriangles(ringVertices(outer, inner), &orion.DrawTrianglesOptions{
		ColorScale: orion.ColorScaleOf(color),
		Transform:  g.toScreen.Translate(pos.XY()),
	})
}

func ringVertices(outer, inner float32) []orion.Vertex2d {
	circumference := outer * math.Pi * 2
	pointCount := int(max(32, min(1024, circumference/16.0)))

	pointAt := func(idx int, radius float32) orion.Vertex2d {
		angle := glm.Rad(float32(idx%pointCount) * float32(math.Pi*2) / float32(pointCount))
		pos := glm.RotationMat3[float32](angle).Transform2(glm.Vec2f{radius, 0})
		return orion.Vertex2d{Position: pos}
	}

	var vertices []orion.Vertex2d
	for idx := range pointCount {
		vertices = append(vertices,
			// first triangle
			pointAt(idx, outer),
			pointAt(idx+1, outer),
			pointAt(idx, inner),

			// second triangle
			pointAt(idx+1, outer),
			pointAt(idx+1, inner),
			pointAt(idx, inner),
		)
	}

	return vertices
}

func (g *Game) spawn(pos glm.Vec2f, vel glm.Vec2f) {
	if g.rng.Float32() > 0.6 {
		return
	}

	g.particles = append(g.particles, Particle{
		Position: pos.Add(randVec(g.rng)),
		Velocity: vel.Add(randVec(g.rng).Scale(4.0)),
	})
}

func (g *Game) fixedTimeStep() (float32, int) {
	// calculate frame delta
	now := time.Now()
	delta := now.Sub(g.lastTime)
	g.lastTime = now
	g.timeStepAcc += delta

	var stepCount int

	step := time.Second / 120
	for g.timeStepAcc > step {
		g.timeStepAcc -= step
		stepCount += 1
	}

	return float32(step.Seconds()), stepCount
}

func (g *Game) sampleNoise(x, y float32) float32 {
	return float32(g.noise.GetNoise2D(fastnoiselite.FNLfloat(x), fastnoiselite.FNLfloat(y)))
}
