package main

import (
	"fmt"
	"image"
	"os"
	"time"

	"math/rand"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	screenmaxX    = 800
	screenmaxY    = 800
	neighbourhood = 70.0
	swarmsize     = 950

	//Adjusts the strength of all the swarming factors together
	f = 1.2

	//Adjusts ruleDontCrowd
	personalSpaceSize  = 16.0
	personalSpacePower = 0.014 * f

	//Adjusts ruleGetClose
	boidTightness = 0.007 * f
	speedCap      = 6.0

	//Adjusts ruleMatchSpeed
	copycatPower = 0.014 * f
)

type boid struct {
	vel   pixel.Vec
	loc   pixel.Vec
	angle float64
}

func (b *boid) move() {
	b.loc = b.loc.Add(b.vel)
	b.angle = b.vel.Angle()
}

func (b *boid) ruleStayOnScreen() {
	if b.loc.X < 0 {
		b.vel.X *= -1
		b.loc.X = 0
	} else if b.loc.X > screenmaxX {
		b.vel.X *= -1
		b.loc.X = screenmaxX
	}
	if b.loc.Y < 0 {
		b.vel.Y *= -1
		b.loc.Y = 0
	} else if b.loc.Y > screenmaxY {
		b.vel.Y *= -1
		b.loc.Y = screenmaxY
	}
}

func (b *boid) ruleGetClose(neighs []*boid) {
	if len(neighs) <= 0 {
		return
	}
	mid := pixel.ZV
	for _, x := range neighs {
		mid = mid.Add(x.loc)
	}
	mid = mid.Scaled(1 / float64(len(neighs)))
	delta := mid.Sub(b.loc).Scaled(boidTightness)
	b.vel = b.vel.Add(delta)
}

func (b *boid) ruleDontCrowd(neighs []*boid) {
	if len(neighs) <= 0 {
		return
	}
	mid := pixel.ZV
	for _, x := range neighs {
		if b.loc.Sub(x.loc).Len() < personalSpaceSize {
			mid = mid.Add(b.loc.Sub(x.loc)) //Head away from crowding neighbour
		}
	}
	b.vel = b.vel.Add(mid.Scaled(personalSpacePower))
}

func (b *boid) ruleMatchVelocity(neighs []*boid) {
	if len(neighs) <= 0 {
		return
	}
	average := pixel.ZV
	for _, x := range neighs {
		average = average.Add(x.vel)
	}
	average = average.Scaled(1 / float64(len(neighs)))
	delta := average.Sub(b.vel).Scaled(copycatPower)
	b.vel = b.vel.Add(delta)
}

func (b *boid) ruleSpeedCap() {
	speed := b.vel.Len()
	if speed > speedCap {
		b.vel = b.vel.Scaled(speedCap / speed)
	}
}

func (b boid) neighbours(swarm *[]boid) []*boid {
	var neigh []*boid
	var dist float64
	for i := 0; i < len(*swarm); i++ {
		dist = b.loc.Sub((*swarm)[i].loc).Len()
		if dist < neighbourhood && b != (*swarm)[i] {
			neigh = append(neigh, &(*swarm)[i])
		}
	}
	return neigh
}

func myRand(n int) float64 {
	return float64(rand.Intn(n))
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Swarm",
		Bounds: pixel.R(0, 0, screenmaxX, screenmaxY),
		VSync:  true, //can be changed to false to test performance
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	win.SetSmooth(true)

	arrowIm, err := loadPictures("../arrow.png")
	if err != nil {
		panic(err)
	}
	batch := pixel.NewBatch(&pixel.TrianglesData{}, arrowIm)

	boidSprite := pixel.NewSprite(arrowIm, arrowIm.Bounds())
	var guys []boid
	var mat pixel.Matrix
	var g *boid

	//Creating the boids
	for i := 0; i < swarmsize; i++ {
		guys = append(guys, boid{pixel.V(myRand(10), myRand(10)), pixel.V(myRand(800), myRand(800)), 0})
	}

	// use these to measure FPS
	var (
		frames = 0
		second = time.Tick(time.Second)
	)

	win.Clear(colornames.Aliceblue)
	for !win.Closed() {
		batch.Clear()

		//Updating boids
		for i := 0; i < len(guys); i++ {
			g = &guys[i]

			//Do swarm behaviour
			neighs := g.neighbours(&guys)
			g.ruleGetClose(neighs)
			g.ruleDontCrowd(neighs)
			g.ruleMatchVelocity(neighs)
			g.ruleStayOnScreen()
			g.ruleSpeedCap()
			g.move()

			mat = pixel.IM.Rotated(pixel.ZV, g.angle).Moved(g.loc)
			boidSprite.Draw(batch, mat)
		}

		win.Clear(colornames.Aliceblue)
		batch.Draw(win)
		win.Update()

		// Tracks and displays FPS
		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
}

//This function is taken from an example in the pixel library
func loadPictures(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func main() {
	pixelgl.Run(run)
}
