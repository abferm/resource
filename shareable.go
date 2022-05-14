package resource

import (
	"context"
	"fmt"
	"sync"
)

func NewShareable[T any](constructor Constructor[T], destructor Destructor[T], cap int) Manager[T] {
	p := &shareable[T]{
		construct: constructor,
		destruct:  destructor,
	}

	return p
}

type shareable[T any] struct {
	construct Constructor[T]
	destruct  Destructor[T]

	h *Handle[T]

	done flag
	l    sync.Mutex
}

func (s *shareable[T]) Acquire(ctx context.Context) (*Handle[T], error) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.done.check() {
		return nil, ErrManagerClosed
	}

	if s.h != nil {
		return s.h, nil
	}

	r, err := s.construct(ctx)
	if err != nil {
		return nil, err
	}
	h := &Handle[T]{
		Resource: r,
		manager:  s,
	}
	s.h = h
	return h, nil
}

func (s *shareable[T]) Release(h *Handle[T]) error {
	// Release is not necessary for shareable resources
	return nil
}

func (s *shareable[T]) Destroy(h *Handle[T]) error {
	s.l.Lock()
	defer s.l.Unlock()
	if h != s.h {
		return fmt.Errorf("can not destroy: %w", ErrInvalidHandle)
	}
	s.h = nil
	return s.destruct(h.Resource)
}

func (p *shareable[T]) Close() error {
	p.l.Lock()
	defer p.l.Unlock()
	if p.done.check() {
		// already closed
		return ErrManagerClosed
	}

	// prevent new acquires
	p.done.set()

	if p.h != nil {
		h := p.h
		p.h = nil
		return p.destruct(h.Resource)
	}

	return nil
}
