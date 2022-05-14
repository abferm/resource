package resource

import (
	"context"
	"fmt"
)

// ErrManagerClosed is returned when an action is attempted on a closed manager
var ErrManagerClosed = fmt.Errorf("manager is closed")

// ErrInvalidHandle is returned when a manager is passed a handle that it isn't
// currently managing. The handle could have been created by another manager, or
// could already have been destroyed.
var ErrInvalidHandle = fmt.Errorf("handle not held by this manager")

// Constructor is a function provided to the Manager for creating new instances
// of the managed resource
type Constructor[T any] func(ctx context.Context) (T, error)

// Destructor is a function provided to the Manager to destroy defective instances
type Destructor[T any] func(T) error

// Handle is a container for a managed resource
type Handle[T any] struct {
	Resource T
	manager  Manager[T]
}

// Release relinquishes control of the resource back to the Manager.
//
// Client code MUST call Release or Destroy when done with a resource.
func (h *Handle[T]) Release() error {
	return h.manager.Release(h)
}

// Destroy informs the Manager that the resource is defective and must
// be destroyed, the Manager is responsible for determining if it should
// attempt to recreate the resource or not.
//
// Client code MUST call Release or Destroy when done with a resource.
func (h *Handle[T]) Destroy() error {
	return h.manager.Destroy(h)
}

type Manager[T any] interface {
	// Acquire attempts to create/checkout the managed resource
	Acquire(ctx context.Context) (*Handle[T], error)
	// Release relinquishes the managed resource back into the Manager
	Release(*Handle[T]) error
	// Destroy destroys a defective resource
	Destroy(*Handle[T]) error
	// Close cleans up all resources in the manager and prevents further calls to Acquire
	Close() error
}
