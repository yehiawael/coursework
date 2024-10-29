package gol

import "uk.ac.bris.cs/gameoflife/util"

var CalculateNext = "GameOfLifeOperations.NextState"

type Response struct {
	NewWorld        [][]byte
	AliveCellsCount []util.Cell
}

type Request struct {
	P     Params
	World [][]byte
}
