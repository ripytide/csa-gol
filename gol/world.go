package gol

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"uk.ac.bris.cs/gameoflife/util"
)

type Dimensions struct {
	width  int
	height int
}

// start and end are inclusive
type Range struct {
	start int
	end   int
}

type World struct {
	world      [][]byte
	dimensions Dimensions
}

func newWorld(dimensions Dimensions) World {
	world := make([][]byte, dimensions.height)
	for i := range world {
		world[i] = make([]byte, dimensions.width)
	}

	return World{world, dimensions}
}

func (world World) processOneTurnWithThreads(newWorld World, threads int) {
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		range_x := Range{start: 0, end: world.dimensions.width}
		range_y := get_sliced_range(i, threads, world.dimensions.height)

		wg.Add(1)
		go func() {
			defer wg.Done()
			world.partialProcessOneTurn(newWorld, range_x, range_y)
		}()
	}

	wg.Wait()
}

func (world World) partialProcessOneTurn(newWorld World, range_x, range_y Range) {
	for y := range_y.start; y < range_y.end; y++ {
		for x := range_x.start; x < range_x.end; x++ {
			world.update_cell(newWorld, x, y)
		}
	}
}

func (world World) update_cell(newWorld World, x int, y int) {
	neighbors := world.countNeigbours(x, y)
	if world.world[y][x] != 0 {
		if neighbors < 2 || neighbors > 3 {
			newWorld.world[y][x] = 0
		} else {
			newWorld.world[y][x] = world.world[y][x]
		}
	} else {
		if neighbors == 3 {
			newWorld.world[y][x] = 255
		} else {
			newWorld.world[y][x] = 0
		}
	}
}

func (world World) to_cells() []util.Cell {
	cells := make([]util.Cell, 0)
	for y := 0; y < world.dimensions.height; y++ {
		for x := 0; x < world.dimensions.width; x++ {
			if world.world[y][x] != 0 {
				cells = append(cells, util.Cell{X: x, Y: y})
			}
		}
	}

	return cells
}

func (world World) countNeigbours(x int, y int) int {
	neighbors := 0

	wrapY := func(v int) int { return wrap(v, world.dimensions.height) }
	wrapX := func(v int) int { return wrap(v, world.dimensions.width) }

	//reads from top-left to bottom-right
	if world.world[wrapY(y+1)][wrapX(x-1)] != 0 {
		neighbors++
	}
	if world.world[wrapY(y+1)][x] != 0 {
		neighbors++
	}
	if world.world[wrapY(y+1)][wrapX(x+1)] != 0 {
		neighbors++
	}
	if world.world[y][wrapX(x-1)] != 0 {
		neighbors++
	}
	if world.world[y][wrapX(x+1)] != 0 {
		neighbors++
	}
	if world.world[wrapY(y-1)][wrapX(x-1)] != 0 {
		neighbors++
	}
	if world.world[wrapY(y-1)][x] != 0 {
		neighbors++
	}
	if world.world[wrapY(y-1)][wrapX(x+1)] != 0 {
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

// writePgmImage receives an array of bytes and writes it to a pgm file.
func (world World) writePgmImage(filename string) {
	_ = os.Mkdir("out", os.ModePerm)

	file, ioError := os.Create(filename)
	util.Check(ioError)
	defer file.Close()

	_, _ = file.WriteString("P5\n")
	_, _ = file.WriteString(strconv.Itoa(world.dimensions.width))
	_, _ = file.WriteString(" ")
	_, _ = file.WriteString(strconv.Itoa(world.dimensions.height))
	_, _ = file.WriteString("\n")
	_, _ = file.WriteString(strconv.Itoa(255))
	_, _ = file.WriteString("\n")

	for y := 0; y < world.dimensions.height; y++ {
		for x := 0; x < world.dimensions.width; x++ {
			_, ioError = file.Write([]byte{world.world[y][x]})
			util.Check(ioError)
		}
	}

	ioError = file.Sync()
	util.Check(ioError)
}

// readPgmImage opens a pgm file returns that file as a 2d byte array
func readPgmImage(expected_dimensions Dimensions) World {

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
		world.world[i/height][i%width] = cell
	}

	return world
}
