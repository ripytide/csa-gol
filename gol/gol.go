package gol

import "time"

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

	active_world := readPgmImage(dimensions)
	other_world := newWorld(dimensions)

	ticker := time.NewTicker(20 * time.Millisecond)

	for i := 0; i < p.Turns; i++ {
		select {
		case <-ticker.C:
			//send the number of cells alive currently
			CellsCount := len(active_world.to_cells())
			events <- AliveCellsCount{CompletedTurns: i, CellsCount: CellsCount}
		default:
		}

		//do a turn
		active_world.processOneTurnWithThreads(other_world, p.Threads)
		//swap active and other
		temp := active_world
		active_world = other_world
		other_world = temp

		events <- TurnComplete{CompletedTurns: i + 1}
	}

	events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: active_world.to_cells()}

	close(events)
}
