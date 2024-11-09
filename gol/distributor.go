package gol

import "C"
import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

var worldMu sync.RWMutex // Protect access to the world state

func saveBoard(p Params, turn int, world [][]byte, c distributorChannels, filename string) {
	filename = fmt.Sprint(filename, "x", turn)
	c.ioCommand <- ioOutput
	c.ioFilename <- filename
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			c.ioOutput <- world[i][j]
		}
	}
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- ImageOutputComplete{
		CompletedTurns: turn,
		Filename:       filename,
	}

}
func handleKeypresses(client *rpc.Client, p Params, res *Response, c distributorChannels, keypresses <-chan rune) {
	filename := fmt.Sprintf("%dx%d", p.ImageWidth, p.ImageHeight)
	paused := false
	//turn := res.Turns
	turn := 0
	for key := range keypresses { // Loop over received keypress events
		switch key {
		case 's': // Save current state
			fmt.Println("Saving current state...")
			worldMu.RLock()
			saveBoard(p, res.Turns, res.NewWorld, c, filename)
			worldMu.RUnlock()
		case 'q': // Quit after saving
			fmt.Println("Quitting after completing current turn...")
			worldMu.RLock()
			saveBoard(p, res.Turns, res.NewWorld, c, filename)
			worldMu.RUnlock()
			// Close keypresses channel to stop the loop
			//close(keypresses)
			// gracefully close the client
			client.Close()
			return
		case 'k': // Quit after saving
			fmt.Println("Quitting after k is pressed")
			worldMu.RLock()
			saveBoard(p, res.Turns, res.NewWorld, c, filename)
			worldMu.RUnlock()

			err := client.Call(Kill, Request{}, new(Response))
			if err != nil {
				fmt.Println("Error when calling Kill", err)
			}

			// Close keypresses channel to stop the loop
			//close(keypresses)
			// gracefully close the client
			client.Close()

			return
		case 'p': // Pause or resume
			fmt.Println("Pausing...")
			paused = true
			//c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
			c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
			err := client.Call(Stop, Request{}, new(Response))
			if err != nil {
				fmt.Println("Error resuming:", err)
			}
			// Wait for another 'p' to resume
			for paused {
				key := <-keypresses //waiting for the new key
				if key == 'p' {
					fmt.Println("Resuming...")
					paused = false
					err := client.Call(Continue, Request{}, new(Response))
					if err != nil {
						fmt.Println("Error resuming:", err)
					}
					c.events <- StateChange{CompletedTurns: turn, NewState: Executing}

				} else if key == 's' {
					// Handle save while paused
					fmt.Println("Saving current state...")
					worldMu.RLock()
					saveBoard(p, res.Turns, res.NewWorld, c, filename)
					worldMu.RUnlock()
				} else if key == 'q' {
					// Handle quit while paused
					fmt.Println("Quitting after completing current turn...")
					worldMu.RLock()
					saveBoard(p, res.Turns, res.NewWorld, c, filename)
					worldMu.RUnlock()
					// Close keypresses channel to stop the loop
					//close(keypresses)
					// gracefully close the client
					client.Close()
					return
				}
			}
		}
	}
}

func makeCall(client *rpc.Client, p Params, world [][]byte, c distributorChannels, keypresses <-chan rune) Response {
	request := Request{P: p, World: world}
	response := new(Response)

	// Initialize response.NewWorld with correct dimensions if it's empty
	if len(response.NewWorld) == 0 {
		response.NewWorld = make([][]byte, p.ImageHeight)
		for i := range response.NewWorld {
			response.NewWorld[i] = make([]byte, p.ImageWidth)
		}
	}

	ticker := time.NewTicker(2 * time.Second) // Runs every 2 second
	defer ticker.Stop()
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				//pause.Lock()
				req := Request{P: p, World: world}
				//fmt.Println("tick")
				//fmt.Println(request.P.Turns)
				//fmt.Println("before Alivecount call")
				err := client.Call(AliveCount, req, response)
				if err != nil {
					fmt.Println("Error when calling", err)
				}
				//fmt.Println("before accessing respone.AliveCellsCount")
				aliveCellsCount := response.AliveCellsCount
				//fmt.Println("before accessing response.Turns")
				completedTurns := response.Turns
				//fmt.Println("before sending the event")
				//fmt.Println("aliveCellsCount:", aliveCellsCount)
				//fmt.Println("completedTurns:", completedTurns)

				// Send an event only if we've completed at least one turn
				if completedTurns > 0 {
					c.events <- AliveCellsCount{
						CompletedTurns: completedTurns,
						CellsCount:     aliveCellsCount,
					}

					fmt.Printf("after sending the event")
				}
				//pause.Unlock()
			case <-done:
				//fmt.Println("Done now")
				return
			}
		}
	}()
	go handleKeypresses(client, p, response, c, keypresses)
	turn := response.Turns
	c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
	err := client.Call(CalculateNext, request, response)
	if err != nil {
		fmt.Println("Error when calling", err)
	}
	done <- true

	//fmt.Println("finished client calling")
	return *response
}

var server = flag.String("server", "127.0.0.1:8031", "IP:port string to connect to as server")

func RunClient(p Params, world [][]byte, c distributorChannels, keypresses <-chan rune) [][]byte {
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
	Res := makeCall(client, p, world, c, keypresses)
	fmt.Println("recieved")
	worldMu.Lock()
	world = Res.NewWorld
	worldMu.Unlock()

	//fmt.Printf("Recieved world of size: %d\n", len(world))

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          Res.AliveCells,
	}

	return world
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keypresses <-chan rune) {

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
	world = RunClient(p, world, c, keypresses)
	saveBoard(p, p.Turns, world, c, filename)
	// TODO: send output back to IO
	c.ioCommand <- ioOutput
	c.ioFilename <- filename

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
