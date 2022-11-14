package gol

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

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

type Dimensions struct {
	width  int
	height int
}

// start and end are inclusive
type Range struct {
	start int
	end   int
}

type World [][]byte

func newWorld(dimensions Dimensions) World {
	world := make(World, dimensions.height)
	for i := range world {
		world[i] = make([]byte, dimensions.width)
	}

	return world
}

// Run starts the processing of Game of Life.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	dimensions := Dimensions{width: p.ImageWidth, height: p.ImageHeight}

	world1 := readPgmImage(dimensions)
	world2 := newWorld(dimensions)

	for i := 0; i < p.Turns; i++ {
		if i%2 == 0 {
			processOneTurnWithThreads(world1, world2, dimensions, p.Threads)
		} else {
			processOneTurnWithThreads(world2, world1, dimensions, p.Threads)
		}
	}
	var finalWorld World
	if p.Turns%2 == 0 {
		finalWorld = world1
	} else {
		finalWorld = world2
	}

	cells := worldToCells(finalWorld, dimensions)
	events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: cells}

	close(events)
}

func processOneTurnWithThreads(oldWorld World, newWorld World, dimensions Dimensions, threads int) {
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		range_x := Range{start: 0, end: dimensions.width}
		range_y := get_sliced_range(i, threads, dimensions.height)

		wg.Add(1)
		go func() {
			defer wg.Done()
			partialProcessOneTurn(oldWorld, newWorld, dimensions, range_x, range_y)
		}()
	}

	wg.Wait()
}

func get_sliced_range(position, threads, length int) Range {
	start := position * length / threads

	var end int
	if position == threads-1 {
		end = length
	} else {
		end = (position + 1) * length / threads
	}

	return Range{start, end}
}

func partialProcessOneTurn(oldWorld World, newWorld World, dimensions Dimensions, range_x, range_y Range) {
	for y := range_y.start; y < range_y.end; y++ {
		for x := range_x.start; x < range_x.end; x++ {
			update_cell(oldWorld, newWorld, x, y, dimensions)
		}
	}
}

func update_cell(oldWorld World, newWorld World, x int, y int, dimensions Dimensions) {
	neighbors := countNeigbours(oldWorld, dimensions, x, y)
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

func worldToCells(world World, dimensions Dimensions) []util.Cell {
	cells := make([]util.Cell, 0)
	for y := 0; y < dimensions.height; y++ {
		for x := 0; x < dimensions.width; x++ {
			if world[y][x] != 0 {
				cells = append(cells, util.Cell{X: x, Y: y})
			}
		}
	}

	return cells
}

func countNeigbours(world World, dimensions Dimensions, x int, y int) int {
	neighbors := 0

	wrapY := func(v int) int { return wrap(v, dimensions.height) }
	wrapX := func(v int) int { return wrap(v, dimensions.width) }

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
func writePgmImage(world World, filename string, dimensions Dimensions) {
	_ = os.Mkdir("out", os.ModePerm)

	file, ioError := os.Create("out/" + filename + ".pgm")
	util.Check(ioError)
	defer file.Close()

	_, _ = file.WriteString("P5\n")
	_, _ = file.WriteString(strconv.Itoa(dimensions.width))
	_, _ = file.WriteString(" ")
	_, _ = file.WriteString(strconv.Itoa(dimensions.height))
	_, _ = file.WriteString("\n")
	_, _ = file.WriteString(strconv.Itoa(255))
	_, _ = file.WriteString("\n")

	for y := 0; y < dimensions.height; y++ {
		for x := 0; x < dimensions.width; x++ {
			_, ioError = file.Write([]byte{world[y][x]})
			util.Check(ioError)
		}
	}

	ioError = file.Sync()
	util.Check(ioError)
}

// readPgmImage opens a pgm file returns that file as a 2d byte array
func readPgmImage(expected_dimensions Dimensions) [][]byte {

	data, ioError := ioutil.ReadFile("images/" + strconv.Itoa(expected_dimensions.width) + "x" + strconv.Itoa(expected_dimensions.height) + ".pgm")
	util.Check(ioError)

	fields := strings.Fields(string(data))

	if fields[0] != "P5" {
		panic("Not a pgm file")
	}

	width, _ := strconv.Atoi(fields[1])
	if width != expected_dimensions.width {
		panic("Incorrect width")
	}

	height, _ := strconv.Atoi(fields[2])
	if height != expected_dimensions.height {
		panic("Incorrect height")
	}

	maxval, _ := strconv.Atoi(fields[3])
	if maxval != 255 {
		panic("Incorrect maxval/bit depth")
	}

	image := []byte(fields[4])

	world := newWorld(expected_dimensions)

	for i, cell := range image {
		world[i/height][i%width] = cell
	}

	return world
}
