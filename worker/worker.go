package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

func calculateNextState(p gol.Params, world [][]byte, topHalo []byte, bottomHalo []byte) [][]byte {
	IMHT := len(world)
	IMWD := len(world[0])

	// Create a new board to store the next state
	newWorld := make([][]byte, IMHT)
	for i := range newWorld {
		newWorld[i] = make([]byte, IMWD)
	}
	fmt.Println("in halo calculatenextstate")
	fmt.Printf("tophalo: %v\n", topHalo)
	for y := 0; y < IMHT; y++ {
		for x := 0; x < IMWD; x++ {

			var sum byte
			// big problem here
			// Calculate neighbors for the first row using the topHalo, if available

			if y == 0 && topHalo != nil {
				//	fmt.Println("iteration number: ", x)
				sum += topHalo[(x-1+IMWD)%IMWD] / 255
				sum += topHalo[x] / 255
				sum += topHalo[(x+1)%IMWD] / 255

				// Calculate the rest of the neighbors from within the sub-grid
				sum += world[y][(x+IMWD-1)%IMWD] / 255
				sum += world[y][(x+1)%IMWD] / 255
				sum += world[(y+1)%IMHT][(x+IMWD-1)%IMWD] / 255
				sum += world[(y+1)%IMHT][x] / 255
				sum += world[(y+1)%IMHT][(x+1)%IMWD] / 255
			} else if y == IMHT-1 && bottomHalo != nil {
				// Calculate neighbors for the last row using the bottomHalo, if available
				sum += bottomHalo[(x-1+IMWD)%IMWD] / 255
				sum += bottomHalo[x] / 255
				sum += bottomHalo[(x+1)%IMWD] / 255

				// Calculate the rest of the neighbors from within the sub-grid
				sum += world[(y+IMHT-1)%IMHT][(x+IMWD-1)%IMWD] / 255
				sum += world[(y+IMHT-1)%IMHT][x] / 255
				sum += world[(y+IMHT-1)%IMHT][(x+1)%IMWD] / 255
				sum += world[y][(x+IMWD-1)%IMWD] / 255
				sum += world[y][(x+1)%IMWD] / 255
			} else {
				// For all other rows (not using halos), calculate all neighbors within the sub-grid itself
				sum += world[(y+IMHT-1)%IMHT][(x+IMWD-1)%IMWD] / 255
				sum += world[(y+IMHT-1)%IMHT][x] / 255
				sum += world[(y+IMHT-1)%IMHT][(x+1)%IMWD] / 255
				sum += world[y][(x+IMWD-1)%IMWD] / 255
				sum += world[y][(x+1)%IMWD] / 255
				sum += world[(y+1)%IMHT][(x+IMWD-1)%IMWD] / 255
				sum += world[(y+1)%IMHT][x] / 255
				sum += world[(y+1)%IMHT][(x+1)%IMWD] / 255
			}

			// Cast the sum to byte for further processing

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
func calculateNextStateEdited(p gol.Params, world [][]byte) [][]byte {

	IMHT := len(world)    // Height of the world
	IMWD := len(world[0]) // Width of the world
	fmt.Println("world legth", len(world))
	// Create a new board to store the next state
	newWorld := make([][]byte, IMHT)
	for i := range newWorld {
		newWorld[i] = make([]byte, IMWD)
	}

	// Copy top and bottom halos from the original world to the new world
	newWorld[0] = world[0]           // Top halo (first row)
	newWorld[IMHT-1] = world[IMHT-1] // Bottom halo (last row)

	for y := 1; y < IMHT-1; y++ {
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

type Worker struct {
	//mu         sync.Mutex
	subGrid    [][]byte
	turns      int
	p          gol.Params
	topHalo    []byte
	bottomHalo []byte
	neighbors  map[string]string
}

func (w *Worker) CombutePart(req gol.WorkerRequest, res *struct{}) error {
	fmt.Println("we recived a part to combute")
	// Only set subGrid and parameters on the first call
	// Set the initial subgrid and parameters only once
	w.subGrid = make([][]byte, len(req.SubGrid))
	for i := range req.SubGrid {
		w.subGrid[i] = make([]byte, len(req.SubGrid[i]))
		w.subGrid[i] = req.SubGrid[i] // Copy each row to ensure dimensions match
	}
	w.subGrid = req.SubGrid
	w.p = req.P

	return nil
}

// RPC method for receiving halo rows from neighbors
func (w *Worker) ReceiveHalo(req gol.HaloRequest, res *gol.WorkerResponse) error {
	//w.mu.Lock()
	//defer w.mu.Unlock()
	fmt.Println("in recieve halo")
	if req.TopHalo != nil {
		w.topHalo = req.TopHalo
	}
	if req.BottomHalo != nil {
		w.bottomHalo = req.BottomHalo
	}

	//w.subGrid = make([][]byte, len(w.subGrid))
	//for i := range w.subGrid {
	//	w.subGrid[i] = make([]byte, len(w.subGrid[0]))
	//}
	fmt.Println("initial w.subgrid", len(w.subGrid))
	newSubGrid := make([][]byte, len(w.subGrid)+2)
	for i := range newSubGrid {
		newSubGrid[i] = make([]byte, len(w.subGrid[0]))
	}

	// Add the top halo as the first row
	newSubGrid[0] = w.topHalo
	// Copy the subgrid in the middle
	copy(newSubGrid[1:len(newSubGrid)-1], w.subGrid)
	// Add the bottom halo as the last row
	newSubGrid[len(newSubGrid)-1] = w.bottomHalo

	//for _, row := range newSubGrid {
	//	fmt.Println(row)
	//}
	fmt.Println("before calculatenext")
	newSubGrid = calculateNextStateEdited(w.p, newSubGrid)

	w.subGrid = newSubGrid[1 : len(newSubGrid)-1]
	fmt.Printf("subGrid has %d rows and %d columns per row\n", len(newSubGrid), len(newSubGrid[0]))
	fmt.Printf("w.subGrid has %d rows and %d columns per row\n", len(w.subGrid), len(w.subGrid[0]))
	res.UpdatedGrid = w.subGrid

	return nil
}

//// Function to exchange halo rows with neighbors directly
//func (w *Worker) exchangeHalos() {
//	// Send top halo to the worker above
//	fmt.Println("in exchange halo")
//	if w.neighbors["top"] != "" {
//		client, err := rpc.Dial("tcp", w.neighbors["top"])
//		if err == nil {
//			req := gol.WorkerandHalo{BottomHalo: w.subGrid[0]}
//			client.Call("Worker.ReceiveHalo", req, nil)
//			client.Close()
//		} else {
//			fmt.Println("failed to send top")
//		}
//	}
//
//	// Send bottom halo to the worker below
//	if w.neighbors["bottom"] != "" {
//		client, err := rpc.Dial("tcp", w.neighbors["bottom"])
//		if err == nil {
//			req := gol.WorkerandHalo{TopHalo: w.subGrid[len(w.subGrid)-1]}
//			client.Call("Worker.ReceiveHalo", req, nil)
//			client.Close()
//		} else {
//			fmt.Println("failed to send bottom")
//		}
//	}
//}

func main() {
	port := flag.String("port", "8032", "Port to listen on")
	flag.Parse()

	fmt.Println("Worker node starting...")

	worker := new(Worker)
	err := rpc.Register(worker)
	if err != nil {
		fmt.Printf("Error registering worker: %v\n", err)
		return
	}

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Printf("Error starting listener: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Worker is listening for RPC calls on port %s\n", *port)
	rpc.Accept(listener)
}
