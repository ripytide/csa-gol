package gol

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"uk.ac.bris.cs/gameoflife/util"
)

// Params provides the details of how to run the Game of Life and which image to load
// from the image/ folder
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type World [][]byte

func newWorld(width int, height int) World {
	world := make(World, height)
	for i := range world {
		world[i] = make([]byte, width)
	}

	return world
}

// Run starts the processing of Game of Life.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	world1 := readPgmImage(p.ImageWidth, p.ImageHeight)
	world2 := newWorld(p.ImageWidth, p.ImageHeight)

	for i := 0; i < p.Turns; i++ {
		if i%2 == 0 {
			world2 = processOneTurn(world1, world2, p.ImageWidth, p.ImageHeight)
		} else {
			world1 = processOneTurn(world2, world1, p.ImageWidth, p.ImageHeight)
		}
	}
	var finalWorld World
	if p.Turns%2 == 0 {
		finalWorld = world1
	} else {
		finalWorld = world2
	}

	cells := worldToCells(finalWorld, p.ImageWidth, p.ImageHeight)
	events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: cells}

	close(events)
}

func processOneTurn(oldWorld World, newWorld World, width int, height int) World {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			neighbors := countNeigbours(oldWorld, x, y, width, height)
			if oldWorld[y][x] != 0 {
				if neighbors < 2 || neighbors > 3 {
					newWorld[y][x] = 0
				} else {
					newWorld[y][x] = oldWorld[y][x]
				}
			} else {
				if neighbors == 3 {
					newWorld[y][x] = 255
				} else {
					newWorld[y][x] = 0
				}
			}
		}
	}

	return newWorld
}

func worldToCells(world World, width int, height int) []util.Cell {
	cells := make([]util.Cell, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if world[y][x] != 0 {
				cells = append(cells, util.Cell{X: x, Y: y})
			}
		}
	}

	return cells
}

func countNeigbours(world World, x int, y int, width int, height int) int {
	neighbors := 0

	wrapY := func(v int) int { return wrap(v, height) }
	wrapX := func(v int) int { return wrap(v, width) }

	//reads from top-left to bottom-right
	if world[wrapY(y+1)][wrapX(x-1)] != 0 {
		neighbors++
	}
	if world[wrapY(y+1)][x] != 0 {
		neighbors++
	}
	if world[wrapY(y+1)][wrapX(x+1)] != 0 {
		neighbors++
	}
	if world[y][wrapX(x-1)] != 0 {
		neighbors++
	}
	if world[y][wrapX(x+1)] != 0 {
		neighbors++
	}
	if world[wrapY(y-1)][wrapX(x-1)] != 0 {
		neighbors++
	}
	if world[wrapY(y-1)][x] != 0 {
		neighbors++
	}
	if world[wrapY(y-1)][wrapX(x+1)] != 0 {
		neighbors++
	}

	return neighbors
}

func wrap(v int, limit int) int {
	if v < 0 {
		return limit - 1
	} else if v >= limit {
		return 0
	} else {
		return v
	}
}

// writePgmImage receives an array of bytes and writes it to a pgm file.
func writePgmImage(world World, filename string, width int, height int) {
	_ = os.Mkdir("out", os.ModePerm)

	file, ioError := os.Create("out/" + filename + ".pgm")
	util.Check(ioError)
	defer file.Close()

	_, _ = file.WriteString("P5\n")
	_, _ = file.WriteString(strconv.Itoa(width))
	_, _ = file.WriteString(" ")
	_, _ = file.WriteString(strconv.Itoa(height))
	_, _ = file.WriteString("\n")
	_, _ = file.WriteString(strconv.Itoa(255))
	_, _ = file.WriteString("\n")

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_, ioError = file.Write([]byte{world[y][x]})
			util.Check(ioError)
		}
	}

	ioError = file.Sync()
	util.Check(ioError)
}

// readPgmImage opens a pgm file returns that file as a 2d byte array
func readPgmImage(expected_width int, expected_height int) [][]byte {

	data, ioError := ioutil.ReadFile("images/" + strconv.Itoa(expected_width) + "x" + strconv.Itoa(expected_height) + ".pgm")
	util.Check(ioError)

	fields := strings.Fields(string(data))

	if fields[0] != "P5" {
		panic("Not a pgm file")
	}

	width, _ := strconv.Atoi(fields[1])
	if width != expected_width {
		panic("Incorrect width")
	}

	height, _ := strconv.Atoi(fields[2])
	if height != expected_height {
		panic("Incorrect height")
	}

	maxval, _ := strconv.Atoi(fields[3])
	if maxval != 255 {
		panic("Incorrect maxval/bit depth")
	}

	image := []byte(fields[4])

	world := newWorld(width, height)

	for i, cell := range image {
		world[i/height][i%width] = cell
	}

	return world
}
