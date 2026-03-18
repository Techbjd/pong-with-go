package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"unsafe"

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
	{1, 1, 1,
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

func (b *ball) update(left *paddle, right *paddle) {
	b.x += b.xv
	b.y += b.yv

	// Top/bottom bounce
	if b.y-float32(b.radius) <= 0 || b.y+float32(b.radius) >= float32(winHeight) {
		b.yv = -b.yv
	}

	// Score
	if b.x < 0 {
		right.score++
		b.pos = getCenter()
	} else if b.x > float32(winWidth) {
		left.score++
		b.pos = getCenter()
	}

	// Paddle collision (left)
	if b.x-float32(b.radius) <= left.x+left.w/2 &&
		b.y >= left.y-left.h/2 &&
		b.y <= left.y+left.h/2 {

		b.xv = float32(rand.Intn(6) + 5)
		b.yv = float32(rand.Intn(6) - 3)
	}

	// Paddle collision (right)
	if b.x+float32(b.radius) >= right.x-right.w/2 &&
		b.y >= right.y-right.h/2 &&
		b.y <= right.y+right.h/2 {

		b.xv = -float32(rand.Intn(6) + 5)
		b.yv = float32(rand.Intn(6) - 3)
	}
}

const paddleSpeed float32 = 400

func (p *paddle) update(keyState []uint8, controllerAxis int16, elapsedTime float32) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		p.y -= paddleSpeed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		p.y += paddleSpeed * elapsedTime
	}

	// Controller
	if math.Abs(float64(controllerAxis)) > 1500 {
		pct := float32(controllerAxis) / 32767.0
		p.y += p.speed * pct * elapsedTime
	}

	// Clamp
	if p.y < p.h/2 {
		p.y = p.h / 2
	}
	if p.y > float32(winHeight)-p.h/2 {
		p.y = float32(winHeight) - p.h/2
	}
}

func (p *paddle) aiupdate(b *ball) {
	p.y = b.y
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

	pixels := make([]byte, winWidth*winHeight*4)

	player1 := paddle{pos{50, 300}, 20, 100, 0, 400, color{255, 255, 255}}
	player2 := paddle{pos{750, 300}, 20, 100, 0, 400, color{255, 255, 255}}
	b := ball{pos{400, 300}, 10, 5, 3, color{255, 255, 255}}

	keyState := sdl.GetKeyboardState()
	var controllerAxis int16

	var lastTime = time.Now()

	for {
		// delta time
		now := time.Now()
		elapsedTime := float32(now.Sub(lastTime).Seconds())
		lastTime = now

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		clear(pixels)

		if state == play {
			player1.update(keyState, controllerAxis, elapsedTime)
			player2.aiupdate(&b)
			b.update(&player1, &player2)
		} else {
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				state = play
			}
		}

		player1.draw(pixels)
		player2.draw(pixels)
		b.draw(pixels)

		tex.Update(nil, unsafe.Pointer(&pixels[0]), winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		sdl.Delay(16)
	}
}