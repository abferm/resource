package resource

import (
	"context"
	"fmt"
	"sync"
)

// NewSingleton makes a manager for a resource of which there should only ever be one instance
func NewSingleton[T any](constructor Constructor[T], destructor Destructor[T]) Manager[T] {
	return NewPool(constructor, destructor, 1)
}

// NewPool makes a manager for a multi-instance resource. This manager will create new resources
// as necessary up to the specified capacity
func NewPool[T any](constructor Constructor[T], destructor Destructor[T], cap int) Manager[T] {
	p := &pool[T]{
		construct: constructor,
		destruct:  destructor,
		cap:       cap,

		available: make(chan *Handle[T], cap),
		all:       make(map[*Handle[T]]struct{}, cap),
	}

	return p
}

type pool[T any] struct {
	construct Constructor[T]
	destruct  Destructor[T]

	available chan *Handle[T]

	cap int

	all  map[*Handle[T]]struct{}
	done flag
	l    sync.Mutex
}

func (p *pool[T]) Acquire(ctx context.Context) (*Handle[T], error) {
	if p.done.check() {
		return nil, ErrManagerClosed
	}

	if p.full() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case h, ok := <-p.available:
			if !ok {
				return nil, ErrManagerClosed
			}
			return h, nil
		}
	}
	return p.getOrNew(ctx)
}

func (p *pool[T]) Release(h *Handle[T]) error {
	p.l.Lock()
	defer p.l.Unlock()
	if p.done.check() {
		// all resources should already be cleaned up
		return nil
	}
	_, ok := p.all[h]
	if !ok {
		return fmt.Errorf("can not release: %w", ErrInvalidHandle)
	}
	p.available <- h
	return nil
}

func (p *pool[T]) Destroy(h *Handle[T]) error {
	p.l.Lock()
	defer p.l.Unlock()
	if p.done.check() {
		// all resources should already be cleaned up
		return nil
	}
	_, ok := p.all[h]
	if !ok {
		return fmt.Errorf("can not destroy: %w", ErrInvalidHandle)
	}
	delete(p.all, h)
	return p.destruct(h.Resource)
}

func (p *pool[T]) Close() error {
	p.l.Lock()
	defer p.l.Unlock()
	if p.done.check() {
		// already closed
		return ErrManagerClosed
	}

	// prevent new acquires
	p.done.set()
	close(p.available)

	// drain available resources
	for range p.available {
	}

	// destroy resources
	for h := range p.all {
		p.destruct(h.Resource)
	}
	p.all = nil

	return nil
}

func (p *pool[T]) getOrNew(ctx context.Context) (*Handle[T], error) {
	p.l.Lock()
	defer p.l.Unlock()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case h, ok := <-p.available:
		if !ok {
			return nil, ErrManagerClosed
		}
		return h, nil
	default:
		h, err := p.new(ctx)
		return h, err
	}
}

func (p *pool[T]) new(ctx context.Context) (*Handle[T], error) {
	r, err := p.construct(ctx)
	if err != nil {
		return nil, err
	}
	h := &Handle[T]{
		Resource: r,
		manager:  p,
	}
	p.all[h] = struct{}{}
	return h, nil
}

func (p *pool[T]) full() bool {
	p.l.Lock()
	defer p.l.Unlock()
	full := len(p.all)+len(p.available) >= p.cap
	return full
}
