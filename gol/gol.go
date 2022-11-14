package gol

// Params provides the details of how to run the Game of Life and which image to load
// from the image/ folder
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Run starts the processing of Game of Life.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	dimensions := Dimensions{width: p.ImageWidth, height: p.ImageHeight}

	world1 := readPgmImage(dimensions)
	world2 := newWorld(dimensions)

	for i := 0; i < p.Turns; i++ {
		if i%2 == 0 {
			world1.processOneTurnWithThreads(world2, p.Threads)
		} else {
			world2.processOneTurnWithThreads(world1, p.Threads)
		}
	}
	var finalWorld World
	if p.Turns%2 == 0 {
		finalWorld = world1
	} else {
		finalWorld = world2
	}

	cells := finalWorld.to_cells()
	events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: cells}

	close(events)
}
