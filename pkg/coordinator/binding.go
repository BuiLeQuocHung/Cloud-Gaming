package coordinator

import (
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

func (b *Binding) Bind(userConn, workerConn *Connection) {
	pair := &Pair{
		user:   userConn,
		worker: workerConn,
	}

	b.workers[workerConn.id] = pair
	b.users[userConn.id] = pair
}

func (b *Binding) RemoveBinding(id string) *Pair {
	var (
		userID   string
		workerID string
		pair     *Pair
	)

	b.Lock()
	defer b.Unlock()

	if p, ok := b.users[id]; ok {
		pair = p
		userID = p.user.id
		workerID = p.worker.id
	} else if p, ok := b.workers[id]; ok {
		pair = p
		userID = p.user.id
		workerID = p.worker.id
	} else {
		// both worker and user disconnected, only one can call this function
		return nil
	}

	delete(b.workers, workerID)
	delete(b.users, userID)

	return pair
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
