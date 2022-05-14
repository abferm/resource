package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/abferm/resource"
)

func TestShareable(t *testing.T) {
	i := 0
	ctor := func(ctx context.Context) (int, error) {
		i++
		return i, nil
	}
	dtor := func(int) error {
		return nil
	}
	manager := resource.NewShareable(ctor, dtor)

	h1, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// shareable release should do nothing
	err = h1.Release()
	if err != nil {
		t.Fatal(err)
	}

	h2, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("the shareable manager should have given us the same resource again")
	}

	err = h1.Destroy()
	if err != nil {
		t.Fatal(err)
	}

	h3, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if h1 == h3 {
		t.Fatalf("the shareable manager should have created a new resource after destroying the old one")
	}

	err = manager.Destroy(h1)
	if !errors.Is(err, resource.ErrInvalidHandle) {
		t.Fatalf("Destroy with an un-tracked handle should have produced error: %s", resource.ErrInvalidHandle)
	}

	err = manager.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = manager.Acquire(context.Background())
	if !errors.Is(err, resource.ErrManagerClosed) {
		t.Fatalf("Acquire after close should have produced error: %s", resource.ErrManagerClosed)
	}

	err = manager.Close()
	if !errors.Is(err, resource.ErrManagerClosed) {
		t.Fatalf("double close should have produced error: %s", resource.ErrManagerClosed)
	}
}
