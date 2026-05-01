package pool

import "sync"

// Resetter is an interface for types that can reset their internal state.
// Types implementing Resetter are expected to bring themselves to a zero/initial state.
type Resetter interface {
	Reset()
}

// Pool is a generic object pool for values that implement the Resetter interface.
// It wraps sync.Pool and ensures that objects are reset before being returned
// back to the pool.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a new Pool for a specific type.
// The newFn function is used to allocate new objects when the pool is empty.
func New[T Resetter](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get retrieves an object from the pool.
// If the pool is empty, a new object is created using the constructor provided to New.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets the object and returns it back to the pool for reuse.
// The Reset method is always called before storing the object.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
