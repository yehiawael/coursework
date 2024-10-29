package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
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

type GameOfLifeOperations struct{}

func (s *GameOfLifeOperations) NextState(req gol.Request, res *gol.Response) (err error) {

	//if req.Message == "" {
	//	err = errors.New("A message must be specified")
	//	return
	//}

	fmt.Println("Got ImageWidth: ", req.P.ImageWidth, " ", req.P.ImageHeight)
	turn := 0

	res.NewWorld = req.World

	fmt.Println("before loop")

	for ; turn < req.P.Turns; turn++ {

		fmt.Println("turn: ", turn)
		res.NewWorld = calculateNextState(req.P, res.NewWorld)
		fmt.Printf("new world len: %d\n", len(res.NewWorld))

		fmt.Println("send event successfully")
	}
	res.AliveCellsCount = calculateAliveCells(req.P, res.NewWorld)
	fmt.Println("finished!")

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

	//conn, err := listener.Accept()
	//if err != nil {
	//	fmt.Printf("Failed to accept connection: %v\n", err)
	//
	//}

	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		fmt.Printf("Failed to accept connection: %v\n", err)
	//		continue
	//	}
	//	go rpc.ServeConn(conn)
	//}

	// This print statement is optional, as `rpc.Accept()` will block indefinitely.
	fmt.Println("server stopped listening")
}
