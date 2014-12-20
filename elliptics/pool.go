package elliptics

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	KeyError        = errors.New("No key")
	counter  uint64 = 0
)

var Pool = contextPool{pool: make(map[uint64]interface{})}

type contextPool struct {
	mutex sync.Mutex
	pool  map[uint64]interface{}
}

func NextContext() uint64 {
	return atomic.AddUint64(&counter, 1)
}

func (p *contextPool) Store(key uint64, value interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.pool[key] = value
	return nil
}

func (p *contextPool) Get(key uint64) (interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if value, exists := p.pool[key]; exists {
		return value, nil
	}

	return nil, KeyError
}

func (p *contextPool) Delete(key uint64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.pool, key)
}
