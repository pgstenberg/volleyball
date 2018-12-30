package server

import (
	uuid "github.com/satori/go.uuid"
)

type snapshot struct {
	Players map[uuid.UUID]*player
	//LastSequenceNumber map[uuid.UUID]uint32
}

func copySnapshot(snapshot1 *snapshot) *snapshot {
	newSnapshot := snapshot{
		Players: make(map[uuid.UUID]*player, len(snapshot1.Players)),
	}
	for index, element := range snapshot1.Players {
		newSnapshot.Players[index] = element.copy()
	}
	return &newSnapshot
}

func diffSnapshot(snapshot0 *snapshot, snapshot1 *snapshot) *snapshot {

	dSnapshot := snapshot{
		Players: make(map[uuid.UUID]*player),
	}

	for id, p := range snapshot1.Players {
		if p.PosX != snapshot0.Players[id].PosX || p.PosY != snapshot0.Players[id].PosY {
			dSnapshot.Players[id] = p.copy()
		}
	}

	return &dSnapshot
}
