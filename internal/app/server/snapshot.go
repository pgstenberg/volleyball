package server

import (
	uuid "github.com/satori/go.uuid"
)

type snapshot struct {
	players            map[uuid.UUID]*player
	lastSequenceNumber map[uuid.UUID]uint32
}

func copySnapshot(snapshot1 *snapshot) *snapshot {
	newSnapshot := snapshot{
		players:            make(map[uuid.UUID]*player, len(snapshot1.players)),
		lastSequenceNumber: make(map[uuid.UUID]uint32, len(snapshot1.lastSequenceNumber)),
	}
	for index, element := range snapshot1.players {

		newSnapshot.players[index] = &player{
			pos: &vector{x: element.pos.x, y: element.pos.y},
			vel: &vector{x: element.vel.x, y: element.vel.y},
			acc: &vector{x: element.acc.x, y: element.acc.y},
		}
	}
	for index, element := range snapshot1.lastSequenceNumber {
		v := element
		newSnapshot.lastSequenceNumber[index] = v
	}
	return &newSnapshot
}

func diffSnapshot(snapshot0 *snapshot, snapshot1 *snapshot) *snapshot {

	dSnapshot := snapshot{
		players:            make(map[uuid.UUID]*player),
		lastSequenceNumber: make(map[uuid.UUID]uint32),
	}

	for id, p := range snapshot1.players {
		if p.pos.x != snapshot0.players[id].pos.x || p.pos.y != snapshot0.players[id].pos.y {
			dSnapshot.players[id] = &player{
				pos: p.pos,
				vel: p.vel,
				acc: p.acc,
			}
		}
	}

	return &dSnapshot
}
