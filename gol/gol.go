package gol

import (
	"fmt"
	"time"
)

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

	//send initial cell flips
	active_world.sendInitialCellFlips(p.Threads, events)

	ticker := time.NewTicker(20 * time.Millisecond)

	quit := false

	for i := 0; i < p.Turns && !quit; i++ {
		//do a turn
		active_world.processOneTurnWithThreads(other_world, p.Threads, events, i)
		//swap active and other
		temp := active_world
		active_world = other_world
		other_world = temp

		events <- TurnComplete{CompletedTurns: i + 1}

		loopy := true
		for loopy {
			select {
			case <-ticker.C:
				//send the number of cells alive currently
				CellsCount := len(active_world.to_cells())
				events <- AliveCellsCount{CompletedTurns: i, CellsCount: CellsCount}
			case key := <-keyPresses:
				switch key {
				case 's':
					println("Generating Output File with Current State")
					active_world.writePgmImage(out_filename(active_world, i))
				case 'q':
					println("Generating Output File with Current State and terminating")
					active_world.writePgmImage(out_filename(active_world, i))
					quit = true
				case 'p':
					println("Pausing execution on execution of turn: ", i)
					for key := range keyPresses {
						if key == 'p' {
							println("Continuing")
							break
						}
					}
				}
			default:
				loopy = false
			}
		}
	}

	events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: active_world.to_cells()}

	filename := out_filename(active_world, p.Turns)
	active_world.writePgmImage(filename)

	events <- ImageOutputComplete{CompletedTurns: p.Turns, Filename: filename}

	close(events)
}

func out_filename(world World, turns int) string {
	return "out/" + fmt.Sprintf("%vx%vx%v.pgm", world.dimensions.width, world.dimensions.height, turns)
}
