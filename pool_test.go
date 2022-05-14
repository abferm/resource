package resource_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/abferm/resource"
	"golang.org/x/sync/errgroup"
)

func TestPool(t *testing.T) {
	i := 0
	ctor := func(ctx context.Context) (int, error) {
		i++
		return i, nil
	}
	dtor := func(int) error {
		return nil
	}

	manager := resource.NewPool(ctor, dtor, 2)

	h1, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if h1.Resource != 1 {
		t.Fatalf("handler contained unexpected value %d != 1", h1.Resource)
	}

	h2, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if h2.Resource != 2 {
		t.Fatalf("handler contained unexpected value %d != 2", h2.Resource)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	_, err = manager.Acquire(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Should have timed out waiting on resource when all are checked out")
	}

	err = h1.Release()
	if err != nil {
		t.Fatal(err)
	}

	hR1, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hR1.Resource != 1 {
		t.Fatalf("should have reaquired original resource %d != 1", hR1.Resource)
	}

	err = hR1.Destroy()
	if err != nil {
		t.Fatal(err)
	}

	h3, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if h3.Resource != 3 {
		t.Fatalf("handler contained unexpected value %d != 3", h3.Resource)
	}
}

func TestDestroyDuringAcquireNoDeadlock(t *testing.T) {
	i := 0
	ctor := func(ctx context.Context) (int, error) {
		i++
		return i, nil
	}
	dtor := func(int) error {
		return nil
	}

	manager := resource.NewSingleton(ctor, dtor)
	defer manager.Close()
	h1, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		_, err := manager.Acquire(ctx)
		return err
	})

	time.Sleep(time.Millisecond * 500)
	err = h1.Destroy()
	if err != nil {
		t.Fatal(err)
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}
