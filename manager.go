package resource

import (
	"context"
	"fmt"
)

var ErrManagerClosed = fmt.Errorf("manager is closed")
var ErrInvalidHandle = fmt.Errorf("handle not held by this manager")

type Constructor[T any] func(ctx context.Context) (T, error)
type Destructor[T any] func(T) error

// Handle is a container for a managed resource
type Handle[T any] struct {
	Resource T
	manager  Manager[T]
}

// Release relinquishes control of the resource back to the Manager.
// Client code MUST call Release or Destroy when done with a resource.
func (h *Handle[T]) Release() error {
	return h.manager.Release(h)
}

// Destroy informs the Manager that the resource is defective and must
// 	be destroyed, the Manager is responsible for determining if it should
// 	attempt to recreate the resource or not.
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
