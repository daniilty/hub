package hub

import (
	"context"
	"sync"
)

type notification struct{}

// Hub - hub for sending some value to subbed consumers.
type Hub struct {
	mux *sync.RWMutex

	val           Cloner
	queue         chan Cloner
	notifications chan notification
	jobDone       chan notification
	notifyOnce    *sync.Once
}

// NewHub - Hub constructor.
func NewHub() *Hub {
	return &Hub{
		mux:           &sync.RWMutex{},
		queue:         make(chan Cloner),
		notifications: make(chan notification),
		jobDone:       make(chan notification, 1),
		notifyOnce:    &sync.Once{},
	}
}

// Send - send value to hub or override existent.
func (h *Hub) Send(val Cloner) {
	h.mux.Lock()
	h.val = val
	h.mux.Unlock()

	// notify listener if it was the first value
	h.notifyOnce.Do(func() {
		h.jobDone <- notification{}
		close(h.jobDone)
	})
}

// Get - notify hub and get value.
func (h *Hub) Get() Cloner {
	// sub to hub
	h.notifications <- notification{}

	return <-h.queue
}

// Listen - listen to incoming subscribtions.
func (h *Hub) Listen(ctx context.Context) {
	for {
		select {
		case <-h.notifications:
			val, ok := h.cloneVal()
			if !ok {
				// if we don't have value, just wait till someone sends it to us
				<-h.jobDone
				val, _ = h.cloneVal()
			}

			h.queue <- val
		case <-ctx.Done():
			h.stop()

			return
		}
	}
}

func (h *Hub) cloneVal() (Cloner, bool) {
	h.mux.RLock()
	defer h.mux.RUnlock()

	if h.val == nil {
		return nil, false
	}

	clone := h.val.Clone()
	return clone, true
}

func (h *Hub) stop() {
	close(h.queue)
	close(h.notifications)
}
