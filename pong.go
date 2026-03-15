package main

// Experiment! draw some crazy stuff!
// Gist it next week and I'll show it off on stream

import (
	"fmt"
	"math/rand"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

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
	color color
}

func (ball *ball) draw(pixels []byte) {
	//yagni-ya aint gona need it
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			if x*x+y*y < ball.radius*ball.radius {
				setPixel(int(ball.x)+x, int(ball.y)+y, color{255, 255, 255}, pixels)
			}
		}
	}
}

func (paddle *paddle) draw(pixels []byte) {
	//position of the kpong bat start from the center
	startx := int(paddle.x) - int(paddle.w/2)

	starty := int(paddle.y) - int(paddle.h/2)

	for y := 0; y < int(paddle.h); y++ {
		for x := 0; x < int(paddle.w); x++ {
			setPixel(startx+x, starty+y, paddle.color, pixels)
		}
	}

}

func (ball *ball) update(leftpaddle *paddle, rightpaddle *paddle) {

	ball.x += ball.xv
	ball.y += ball.yv

	// Top wall
	if ball.y-float32(ball.radius) <= 0 {

		ball.y = float32(ball.radius)

		ball.yv = float32(rand.Intn(10) - 5)

		ball.xv = float32(rand.Intn(6) + 2)
		if rand.Intn(2) == 0 {
			ball.xv = -ball.xv
		}
	}

	// Bottom wall
	if ball.y+float32(ball.radius) >= float32(winHeight) {
		ball.yv = -ball.yv
		ball.xv = -ball.xv

		if rand.Intn(2) == 0 {
			ball.xv = -ball.xv
		}
		ball.y = float32(winHeight - ball.radius)
	}
	/*
		// Left wall
		if ball.x-float32(ball.radius) <= 0 {
			ball.xv = -ball.xv
			ball.x = float32(ball.radius)
		}

		// Right wall
		if ball.x+float32(ball.radius) >= float32(winWidth) {
			ball.xv = -ball.xv
			ball.x = float32(winWidth - ball.radius)
		}
	*/
	// Score condition
	if ball.x < 0 || ball.x > float32(winWidth) {
		ball.x = 300
		ball.y = 300
	}

	if ball.x-float32(ball.radius) <= leftpaddle.x+float32(leftpaddle.w)/2 &&
		ball.y >= leftpaddle.y-float32(leftpaddle.h)/2 &&
		ball.y <= leftpaddle.y+float32(leftpaddle.h)/2 {

		ball.xv = float32(rand.Intn(6) + 4) // positive → right
	}
	if ball.x+float32(ball.radius) >= rightpaddle.x-float32(rightpaddle.w)/2 &&
		ball.y >= rightpaddle.y-float32(rightpaddle.h)/2 &&
		ball.y <= rightpaddle.y+float32(rightpaddle.h)/2 {

		ball.xv = -float32(rand.Intn(6) + 4) // negative → left
	}

}

const paddleSpeed float32 = 10

func (paddle *paddle) update(keyState []uint8) {
	if keyState[sdl.SCANCODE_UP] != 0 {

		paddle.pos.y -= paddleSpeed
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		paddle.pos.y += paddleSpeed
	}
	// top boundary
	if paddle.pos.y < float32(paddle.h/2) {
		paddle.pos.y = float32(paddle.h / 2)
	}

	// bottom boundary
	if paddle.pos.y > (float32(winHeight) - (paddle.h / 2)) {
		paddle.pos.y = (float32(winHeight) - paddle.h/2)
	}

}

func (paddle *paddle) aiupdate(ball *ball) {
	paddle.y = ball.y
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}

}

func main() {

	// Added after EP06 to address macosx issues
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)

	// /* for y := winHeight; y < winHeight; y++ {
	// 	for x := winWidth; x < winWidth; x++ {
	// 		setPixel(x, y, color{byte(x % 255), byte(y % 255), 0}, pixels)
	// 	}
	// } */

	/* bg := color{255, 255, 255}
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {
			setPixel(x+10, y+30, bg, pixels)
		}
	} */
	clear(pixels)
	player1 := paddle{pos{50, 100}, 20, 100, color{255, 255, 255}}
	player2 := paddle{pos{float32(winWidth) - 50, 100}, 20, 100, color{255, 255, 255}}
	ball := ball{pos{300, 300}, 20, 5, 10, color{255, 255, 255}}

	keyState := sdl.GetKeyboardState()
	rand.Seed(int64(sdl.GetTicks()))
	// Changd after EP 06 to address MacOSX
	// OSX requires that you consume events for windows to open and work properly
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}

		}
		clear(pixels)
		player1.draw(pixels)

		player1.update(keyState)
		player2.draw(pixels)
		player2.aiupdate(&ball)
		ball.draw(pixels)
		ball.update(&player1, &player2)
		tex.Update(nil, unsafe.Pointer(&pixels[0]), winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		sdl.Delay(16)
	}

	// clear(pixels)

}
