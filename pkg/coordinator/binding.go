package coordinator

import (
	"log"
	"math/rand"
	"sync"
)

type (
	Binding struct {
		workers map[string]*Pair
		users   map[string]*Pair
		mu      sync.Mutex
	}

	Pair struct {
		user   *Connection
		worker *Connection
	}
)

func NewBinding() *Binding {
	return &Binding{
		workers: make(map[string]*Pair),
		users:   make(map[string]*Pair),
		mu:      sync.Mutex{},
	}
}

func (c *Coordinator) bindUserAndWorker(userConn *Connection) bool {
	c.binding.Lock()
	defer c.binding.Unlock()

	userId := userConn.id
	log.Println(c.freeWorkers)

	n := len(c.freeWorkers)
	if n == 0 {
		return false
	}

	idx := rand.Int() % n
	workerConn := c.freeWorkers[idx]
	log.Println("random idx: ", idx)

	pair := &Pair{
		user:   userConn,
		worker: workerConn,
	}

	log.Println("pair: ", *pair)

	c.binding.users[userId] = pair
	c.binding.workers[workerConn.id] = pair

	temp := []*Connection{}
	temp = append(temp, c.freeWorkers[:idx]...)
	temp = append(temp, c.freeWorkers[idx+1:]...)
	c.freeWorkers = temp

	log.Println(c.freeWorkers)
	return true
}

func (b *Binding) Lock() {
	b.mu.Lock()
}

func (b *Binding) Unlock() {
	b.mu.Unlock()
}

func (b *Binding) GetPair(id string) *Pair {
	if p, ok := b.users[id]; ok {
		return p
	}

	if p, ok := b.workers[id]; ok {
		return p
	}

	return nil
}

func (b *Binding) IsUserPaired(userId string) bool {
	_, ok := b.users[userId]
	return ok
}

func (b *Binding) IsWorkerPaired(workerId string) bool {
	_, ok := b.workers[workerId]
	return ok
}
