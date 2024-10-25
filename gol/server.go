package gol

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	//	"errors"
	//	"flag"
	//	"fmt"
	//	"net"
	"time"
)

type GameOfLifeOperations struct{}

func (s *GameOfLifeOperations) NextState(req Request, res *Response) (err error) {
	//if req.Message == "" {
	//	err = errors.New("A message must be specified")
	//	return
	//}
	fmt.Println("Got ImageWidth: ", req.p.ImageWidth, " ", req.p.ImageHeight)
	turn := 0
	for ; turn < req.p.Turns; turn++ {
		res.newworld = calculateNextState(req.p, req.world)
		req.c.events <- StateChange{turn, Executing}
	}
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)

}
