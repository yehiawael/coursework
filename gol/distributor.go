package gol

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

//func calculateNextState(p Params, world [][]byte) [][]byte {
//	height := p.ImageHeight
//	width := p.ImageWidth
//	newworld := make([][]byte, height)
//	for i := range newworld {
//		newworld[i] = make([]byte, width)
//	}
//
//	countliveneighbours := func(x int, y int) int {
//		livecount := 0
//		for i := -1; i <= 1; i++ {
//			for j := -1; j <= 1; j++ {
//				if i == 0 && j == 0 {
//					continue
//				}
//				newx := (x + i + height) % height
//				newy := (y + j + width) % width
//				if world[newx][newy] == 255 {
//					livecount++
//				}
//			}
//		}
//		return livecount
//	}
//	for x := 0; x < height; x++ {
//		for y := 0; y < width; y++ {
//			lives := countliveneighbours(x, y)
//			if world[x][y] == 255 {
//				if lives < 2 || lives > 3 {
//					newworld[x][y] = 0
//				}
//				if lives == 2 || lives == 3 {
//					newworld[x][y] = 255
//				}
//			}
//			if world[x][y] == 0 {
//				if lives == 3 {
//					newworld[x][y] = 255
//				}
//			}
//		}
//	}
//	return newworld
//}
//
//
//
//func calculateAliveCells(p Params, world [][]byte) []util.Cell {
//	height := p.ImageHeight
//	width := p.ImageWidth
//	var cells []util.Cell
//	for w := 0; w < height; w++ {
//		for h := 0; h < width; h++ {
//			if world[w][h] == 255 {
//				cells = append(cells, util.Cell{Y: w, X: h})
//			}
//		}
//	}
//	return cells
//}
//

func calculateNextState(p Params, world [][]byte) [][]byte {

	IMHT := p.ImageHeight // Height of the world
	IMWD := p.ImageWidth  // Width of the world

	// Create a new board to store the next state
	newWorld := make([][]byte, IMHT)
	for i := range newWorld {
		newWorld[i] = make([]byte, IMWD)
	}

	for y := 0; y < IMHT; y++ {
		for x := 0; x < IMWD; x++ {
			// Calculate the sum of the neighbors
			sum := world[(y+IMHT-1)%IMHT][(x+IMWD-1)%IMWD]/255 +
				world[(y+IMHT-1)%IMHT][(x+IMWD)%IMWD]/255 +
				world[(y+IMHT-1)%IMHT][(x+IMWD+1)%IMWD]/255 +
				world[(y+IMHT)%IMHT][(x+IMWD-1)%IMWD]/255 +
				world[(y+IMHT)%IMHT][(x+IMWD+1)%IMWD]/255 +
				world[(y+IMHT+1)%IMHT][(x+IMWD-1)%IMWD]/255 +
				world[(y+IMHT+1)%IMHT][(x+IMWD)%IMWD]/255 +
				world[(y+IMHT+1)%IMHT][(x+IMWD+1)%IMWD]/255

			// Update the new world based on the rules
			if world[y][x] == 255 {
				if sum < 2 || sum > 3 {
					newWorld[y][x] = 0
				} else {
					newWorld[y][x] = 255
				}
			} else {
				if sum == 3 {
					newWorld[y][x] = 255
				} else {
					newWorld[y][x] = 0
				}
			}
		}
	}

	return newWorld
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {

	aliveCells := make([]util.Cell, 0, p.ImageHeight)

	for w := 0; w < p.ImageWidth; w++ {
		for h := 0; h < p.ImageHeight; h++ {

			//fmt.Println(world[h][w])
			if world[h][w] == 255 {
				cellCord := util.Cell{X: w, Y: h}
				//fmt.Println(h)
				//fmt.Printf("width : %d", h)

				aliveCells = append(aliveCells, cellCord)
			}
		}

	}
	return aliveCells
}
func makeCall(client *rpc.Client, p Params, world [][]byte, c distributorChannels) {
	request := Request{p: p, world: world, c: c}
	response := new(Response)
	client.Call(CalculateNext, request, response)
	fmt.Println("Responded")
}

func RunClient(p Params, world [][]byte, c distributorChannels) {
	// Define the server address flag (default to your server's IP and port)
	server := flag.String("server", "52.87.172.22:8030", "IP:port string to connect to as server")
	flag.Parse()

	// Establish connection to the server
	client, err := rpc.Dial("tcp", *server)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	// Open the wordlist file using the correct path
	file, err := os.Open("../wordlist")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Scan the file line by line and send each line to the server
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Println("Called: " + t)
		makeCall(client, p, world, c)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from file: %v", err)
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	// Seng ioinout command
	// send filename
	// nested

	filename := fmt.Sprintf("%dx%d", p.ImageWidth, p.ImageHeight)
	c.ioCommand <- ioInput
	c.ioFilename <- filename

	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			eachByte := <-c.ioInput
			world[i][j] = eachByte
		}
	}
	go RunClient(p, world, c)
	// TODO: Execute all turns of the Game of Life.
	//turn := 0
	//for ; turn < p.Turns; turn++ {
	//	world = calculateNextState(p, world)
	//	c.events <- StateChange{turn, Executing}
	//}

	// TODO: send output back to IO
	c.ioCommand <- ioOutput
	c.ioFilename <- filename
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			c.ioOutput <- world[i][j]
		}
	}
	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          calculateAliveCells(p, world),
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	// i changed turn to p.Turns
	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
