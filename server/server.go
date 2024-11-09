package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"

	"time"
)

func calculateNextState(p gol.Params, world [][]byte) [][]byte {

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

func calculateAliveCells(p gol.Params, world [][]byte) []util.Cell {

	aliveCells := make([]util.Cell, 0, p.ImageHeight)

	for w := 0; w < p.ImageWidth; w++ {
		for h := 0; h < p.ImageHeight; h++ {
			if world[h][w] == 255 {
				cellCord := util.Cell{X: w, Y: h}
				aliveCells = append(aliveCells, cellCord)

			}
		}

	}

	return aliveCells
}

//func calculateAliveCells(p gol.Params, world [][]byte) []util.Cell {
//	height := p.ImageHeight
//	width := p.ImageWidth
//	var cells []util.Cell
//	for w := 0; w < height; w++ {
//		for h := 0; h < width; h++ {
//			fmt.Println("inside for loop")
//			if world[w][h] == 255 {
//				cells = append(cells, util.Cell{Y: w, X: h})
//			}
//		}
//	}
//	return cells
//}

type GameOfLifeOperations struct {
	world  [][]byte
	turns  int
	mu     sync.Mutex
	paused bool
}

func (s *GameOfLifeOperations) Paused(req gol.Request, res *gol.Response) (err error) {
	s.mu.Lock()
	fmt.Println("Pausing the server...")
	s.paused = true
	s.mu.Unlock()
	return nil
}

func (s *GameOfLifeOperations) Resume(req gol.Request, res *gol.Response) (err error) {
	s.mu.Lock()
	fmt.Println("Resuming the Server...")
	s.paused = false
	s.mu.Unlock()
	return nil
}

func (s *GameOfLifeOperations) Quit(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("Received kill command. Shutting down server...")

	// Perform any necessary cleanup here
	s.turns = 0
	s.world = nil

	// Exit the program immediately with an exit code of 0 (successful shutdown)
	os.Exit(0)

	// Note: os.Exit() does not return, so any code after this line is unreachable
	return nil
}
func (s *GameOfLifeOperations) AliveCells(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("now in alivecells")
	fmt.Println(req.P.ImageHeight)
	res.AliveCells = calculateAliveCells(req.P, s.world)
	fmt.Println("we calculated alivecells")
	res.AliveCellsCount = len(res.AliveCells)
	res.Turns = s.turns
	fmt.Println("finished alivecells")
	return
}

func (s *GameOfLifeOperations) NextState(req gol.Request, res *gol.Response) (err error) {
	turn := 0
	res.NewWorld = req.World
	for ; turn < req.P.Turns; turn++ {

		paused := s.paused

		//If paused, wait until resumed
		for paused {
			time.Sleep(50 * time.Millisecond) // Check the state periodically so it doesn't abuse the mutex lock
			paused = s.paused
		}
		s.turns = turn
		s.world = calculateNextState(req.P, res.NewWorld)
		res.NewWorld = s.world

	}
	res.AliveCells = calculateAliveCells(req.P, res.NewWorld)

	return
}

func main() {
	fmt.Println("RPC server starting...")

	pAddr := flag.String("port", "8031", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	// Register the RPC server type
	err := rpc.Register(&GameOfLifeOperations{})
	if err != nil {
		fmt.Printf("Failed to register RPC server: %v\n", err)
		return
	}

	// Start listening for incoming connections
	listener, err := net.Listen("tcp", ":"+*pAddr)
	if err != nil {
		fmt.Printf("Failed to start listener: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server is listening on port %s\n", *pAddr)
	rpc.Accept(listener)

	// This print statement is optional, as `rpc.Accept()` will block indefinitely.
	fmt.Println("server stopped listening")
}
