package gol

var CalculateNext = "GameOfLifeOperations.NextState"

type Response struct {
	newworld [][]byte
}

type Request struct {
	p     Params
	world [][]byte
	c     distributorChannels
}
