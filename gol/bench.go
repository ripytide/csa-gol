package gol

import (
	"fmt"
	"os"
	"time"
)

func BenchEverything(filename string) {
	turns := []int{0, 1, 10, 100, 1000, 10000}
	sizes := []int{256, 512}
	threads := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	for _, turn := range turns {
		for _, size := range sizes {
			for _, thread := range threads {
				fmt.Printf("Bench With Turns=%d Size=%d Threads=%d\n", turn, size, thread)

				//initialise the slices
				dimensions := Dimensions{width: size, height: size}
				active_world := readPgmImage(dimensions)
				other_world := newWorld(dimensions)

				//run an individual bench
				start := time.Now()
				bareProcessTurns(active_world, other_world, turn, thread)
				elapsed := time.Since(start)

				//append to result
				f, _ := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				defer f.Close()
				f.WriteString(fmt.Sprintf("%d, %d, %d, %f\n", turn, size, thread, elapsed.Seconds()))
			}
		}
	}
}

func bareProcessTurns(active_world World, other_world World, turns int, threads int) {
	for i := 0; i < turns; i++ {
		//do a bare turn
		active_world.bareProcessOneTurn(other_world, threads)
		//swap active and other
		temp := active_world
		active_world = other_world
		other_world = temp
	}
}
