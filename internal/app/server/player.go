package server

type player struct {
	pos *vector
	vel *vector
	acc *vector
}

type playerInput struct {
	sequenceNumber uint32
	value          uint8
}

func (p *player) proccessInput(value uint8, delta float64) {

	switch value {
	case 1:
		p.acc.set(-40, 0)
	case 2:
		p.acc.set(40, 0)
	}

	p.acc.mul(delta)
	p.vel.add(p.acc)
	p.acc.mul(delta)

	p.pos.add(p.vel)

	///fmt.Printf("X: %d, Y: %d\n", p.pos.x, p.pos.y)

}
