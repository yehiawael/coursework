package gol

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
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

func makeCall(client *rpc.Client, p Params, world [][]byte) Response {
	request := Request{P: p, World: world}
	response := new(Response)
	fmt.Println("client call now")
	err := client.Call(CalculateNext, request, response)
	if err != nil {
		fmt.Println("Error when calling", err)
	}
	fmt.Println("finished client calling")
	return *response
}

var server = flag.String("server", "127.0.0.1:8031", "IP:port string to connect to as server")

func RunClient(p Params, world [][]byte, c distributorChannels) [][]byte {
	// Define the server address flag (default to your server's IP and port)
	//server := flag.String("server", "127.0.0.1:8031", "IP:port string to connect to as server")
	flag.Parse()
	fmt.Println("attempting to dial...")
	// Establish connection to the server
	client, err := rpc.Dial("tcp", *server)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()
	fmt.Println("connected")
	Res := makeCall(client, p, world)
	fmt.Println("recieved")
	world = Res.NewWorld
	fmt.Printf("Recieved world of size: %d\n", len(world))

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          Res.AliveCellsCount,
	}

	return world
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

	world = RunClient(p, world, c)
	// TODO: Execute all turns of the Game of Life.
	//turn := 0
	//for ; turn < p.Turns; turn++ {
	//	world = calculateNextState(p, world)
	//	c.events <- StateChange{turn, Executing}
	//}

	// TODO: send output back to IO
	c.ioCommand <- ioOutput
	c.ioFilename <- filename

	//fmt.Println(world[p]"RPC server starting...")

	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			c.ioOutput <- world[i][j]
		}
	}
	// TODO: Report the final state using FinalTurnCompleteEvent.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	// i changed turn to p.Turns
	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
