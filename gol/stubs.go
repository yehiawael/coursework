package gol

import "uk.ac.bris.cs/gameoflife/util"

//between client and server
var CalculateNext = "GameOfLifeOperations.NextState"
var AliveCount = "GameOfLifeOperations.AliveCells"
var Kill = "GameOfLifeOperations.Quit"
var Stop = "GameOfLifeOperations.Paused"
var Continue = "GameOfLifeOperations.Resume"

//var Server = "GameOfLifeOperations.Broker"

type Response struct {
	NewWorld        [][]byte
	AliveCells      []util.Cell
	AliveCellsCount int
	Turns           int
}

type Request struct {
	P     Params
	World [][]byte
}

////between server and worker
//var Combute = "Worker.CombutePart"
//var Halo = "Worker.ReceiveHalo"
//
//type WorkerRequest struct {
//	SubGrid [][]byte
//	P       Params
//}
//
//type WorkerResponse struct {
//	UpdatedGrid [][]byte
//	TopHalo     []byte
//	BottomHalo  []byte
//}
//
////between workers
//type HaloRequest struct {
//	TopHalo    []byte
//	BottomHalo []byte
//}
