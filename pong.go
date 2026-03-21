package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"unsafe"

	noise "github.com/Techbjd/pong/Noise"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

type gameState int

const (
	start gameState = iota
	play
)

var state = start

var nums = [][]byte{
	{
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{
		1, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		0, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
}

type color struct {
	r, g, b byte
}

type pos struct {
	x, y float32
}

type ball struct {
	pos
	radius int
	xv     float32
	yv     float32
	color  color
}

type paddle struct {
	pos
	w     float32
	h     float32
	score int
	speed float32
	color color
}

func lerpc(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 color, pct float32) color {
	return color{lerpc(c1.r, c2.r, pct), lerpc(c1.g, c2.g, pct), lerpc(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleandDraw(noise []float32, min, max float32, gradient []color, w, h int) []byte {
	result := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale

	for i := range noise {
		value := noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(value))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
		result[p+3] = 255
	}
	return result
}

func getDualGradientColor(c1, c2, c3, c4 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / 255.0

		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*2)
		} else {
			result[i] = colorLerp(c3, c4, (pct-0.5)*2)
		}
	}
	return result
}

func drawNumber(pos pos, color color, size int, num int, pixels []byte) {
	startX := int(pos.x) - (size*3)/2
	startY := int(pos.y) - (size*5)/2

	for i, v := range nums[num] {
		if v == 1 {
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixel(x, y, color, pixels)
				}
			}
		}
		startX += size
		if (i+1)%3 == 0 {
			startY += size
			startX -= size * 3
		}
	}
}

func (b *ball) draw(pixels []byte) {
	for y := -b.radius; y < b.radius; y++ {
		for x := -b.radius; x < b.radius; x++ {
			if x*x+y*y < b.radius*b.radius {
				setPixel(int(b.x)+x, int(b.y)+y, b.color, pixels)
			}
		}
	}
}

func lerp(a float32, b float32, pct float32) float32 {
	return a + pct*(b-a)
}

func (p *paddle) draw(pixels []byte) {
	startx := int(p.x) - int(p.w/2)
	starty := int(p.y) - int(p.h/2)

	for y := 0; y < int(p.h); y++ {
		for x := 0; x < int(p.w); x++ {
			setPixel(startx+x, starty+y, p.color, pixels)
		}
	}

	numX := lerp(p.pos.x, getCenter().x, 0.2)
	drawNumber(pos{numX, 35}, p.color, 10, p.score, pixels)
}

func (b *ball) update(left *paddle, right *paddle, elapsedTime float32) {
	// ✅ FIXED
	b.x += b.xv * elapsedTime
	b.y += b.yv * elapsedTime

	if b.y-float32(b.radius) <= 0 || b.y+float32(b.radius) >= float32(winHeight) {
		b.yv = -b.yv
	}

	if b.x < 0 {
		right.score++
		b.pos = getCenter()
	} else if b.x > float32(winWidth) {
		left.score++
		b.pos = getCenter()
	}

	if b.x-float32(b.radius) <= left.x+left.w/2 &&
		b.y >= left.y-left.h/2 &&
		b.y <= left.y+left.h/2 {

		b.xv = float32(rand.Intn(200) + 300)
		b.yv = float32(rand.Intn(200) - 100)
	}

	if b.x+float32(b.radius) >= right.x-right.w/2 &&
		b.y >= right.y-right.h/2 &&
		b.y <= right.y+right.h/2 {

		b.xv = -float32(rand.Intn(200) + 300)
		b.yv = float32(rand.Intn(200) - 100)
	}
}

func (p *paddle) update(keyState []uint8, controllerAxis int16, elapsedTime float32) {
	const paddleSpeed float32 = 400
	if keyState[sdl.SCANCODE_UP] != 0 {
		p.y -= paddleSpeed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		p.y += paddleSpeed * elapsedTime
	}

	if math.Abs(float64(controllerAxis)) > 1500 {
		pct := float32(controllerAxis) / 32767.0
		p.y += p.speed * pct * elapsedTime
	}

	if p.y < p.h/2 {
		p.y = p.h / 2
	}
	if p.y > float32(winHeight)-p.h/2 {
		p.y = float32(winHeight) - p.h/2
	}
}

func (p *paddle) aiupdate(b *ball, elapsedTime float32) {
	// ✅ FIXED
	p.y += (b.y - p.y) * 10 * elapsedTime
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c color, pixels []byte) {
	if x < 0 || x >= winWidth || y < 0 || y >= winHeight {
		return
	}
	index := (y*winWidth + x) * 4
	pixels[index] = c.r
	pixels[index+1] = c.g
	pixels[index+2] = c.b
	pixels[index+3] = 255
}

func getCenter() pos {
	return pos{float32(winWidth) / 2, float32(winHeight) / 2}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, _ := sdl.CreateWindow("Pong", 0, 0, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	defer window.Destroy()

	renderer, _ := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	defer renderer.Destroy()

	tex, _ := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		int32(winWidth), int32(winHeight),
	)
	defer tex.Destroy()

	var ControllerHandlers []*sdl.GameController
	for i := 0; i < sdl.NumJoysticks(); i++ {
		ControllerHandlers = append(ControllerHandlers, sdl.GameControllerOpen(i))
		defer ControllerHandlers[i].Close()
	}

	pixels := make([]byte, winWidth*winHeight*4)

	player1 := paddle{pos{50, 300}, 20, 100, 0, 400, color{255, 255, 255}}
	player2 := paddle{pos{750, 300}, 20, 100, 0, 400, color{255, 255, 255}}
	b := ball{pos{400, 300}, 10, 300, 180, color{255, 255, 255}}

	keyState := sdl.GetKeyboardState()
	var controllerAxis int16
	noise, min, max := noise.MakeNoise(noise.FBM, .001, 0.5, 2, 3, float32(winWidth), float32(winHeight))
	fmt.Println(noise, min, max)
	gradient := getGradient(color{255, 255, 0}, color{255, 0, 0})
	noisePixels := rescaleandDraw(noise, min, max, gradient, winWidth, winHeight)

	var lastTime = time.Now()

	for {
		now := time.Now()
		elapsedTime := float32(now.Sub(lastTime).Seconds())
		lastTime = now

		if elapsedTime > 0.05 {
			elapsedTime = 0.05
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		for _, controller := range ControllerHandlers {
			if controller != nil {
				controllerAxis = controller.Joystick().Axis(sdl.CONTROLLER_AXIS_LEFTY)
			}
		}

		if state == play {
			player1.update(keyState, controllerAxis, elapsedTime)
			player2.aiupdate(&b, elapsedTime)
			b.update(&player1, &player2, elapsedTime)
		} else {
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				if player1.score == 3 || player2.score == 3 {
					player1.score = 0
					player2.score = 0
				}
				state = play
			}
		}

		for i := range noisePixels {
			pixels[i] = noisePixels[i]

		}

		player1.draw(pixels)
		player2.draw(pixels)
		b.draw(pixels)

		tex.Update(nil, unsafe.Pointer(&pixels[0]), winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		elapsedTime = float32(time.Since(now).Seconds())
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime*1000.0))
			elapsedTime = float32(time.Since(now).Seconds())
		}
	}
}
