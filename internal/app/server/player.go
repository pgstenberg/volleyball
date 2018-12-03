package server

type Player struct {
	Id uint8
	X int
	Y int
}

type PlayerInput struct {
	Id uint8
	SequenceNumber uint32
	Value uint8
}
