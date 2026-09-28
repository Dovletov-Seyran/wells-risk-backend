package auth

import "sync"

const currentPhysicianID = 1

type CurrentPhysician struct {
	PhysicianID int
}

var (
	once     sync.Once
	instance *CurrentPhysician
)

func Current() *CurrentPhysician {
	once.Do(func() {
		instance = &CurrentPhysician{PhysicianID: currentPhysicianID}
	})

	return instance
}
