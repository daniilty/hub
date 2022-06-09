package hub_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/daniilty/hub"
)

type dummyCloner struct {
	val uint32
}

func (d *dummyCloner) Clone() hub.Cloner {
	return &dummyCloner{
		val: d.val,
	}
}

func TestHub(t *testing.T) {
	var res uint32

	hub := hub.NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	hWG := &sync.WaitGroup{}
	hWG.Add(1)

	go func() {
		hub.Listen(ctx)
		hWG.Done()
	}()

	sWG := &sync.WaitGroup{}
	sWG.Add(3)

	go func() {
		val := hub.Get().(*dummyCloner)
		atomic.AddUint32(&res, val.val)
		sWG.Done()
	}()

	go func() {
		val := hub.Get().(*dummyCloner)
		atomic.AddUint32(&res, val.val)
		sWG.Done()
	}()

	go func() {
		val := hub.Get().(*dummyCloner)
		atomic.AddUint32(&res, val.val)
		sWG.Done()
	}()

	hub.Send(&dummyCloner{
		val: 10,
	})

	sWG.Wait()
	cancel()
	hWG.Wait()

	if res != 30 {
		t.Fatal("unexpected result")
	}
}
