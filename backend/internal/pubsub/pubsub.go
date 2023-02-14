package pubsub

import (
	"context"
	"sync"

	"etok.codes/qg/backend/qg"
)

// Publisher implements an event publisher.
type Publisher struct {
	subs sync.Map
}

var (
	_ qg.Publisher  = (*Publisher)(nil)
	_ qg.Subscriber = (*Publisher)(nil)
)

// NewPublisher creates a new publisher.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// Subscribe implements the Subscriber interface.
func (p *Publisher) Subscribe(ctx context.Context, out chan<- qg.Event) error {
	p.subs.Store(out, struct{}{})
	return nil
}

// Unsubscribe implements the Subscriber interface.
func (p *Publisher) Unsubscribe(ctx context.Context, out chan<- qg.Event) error {
	p.subs.Delete(out)
	return nil
}

// Publish implements the Publisher interface.
func (p *Publisher) Publish(ctx context.Context, msg qg.Event) error {
	p.subs.Range(func(key, _ any) bool {
		k := key.(chan<- qg.Event)
		select {
		case <-ctx.Done():
			return false
		case k <- msg:
			return true
		default:
			p.subs.Delete(k)
			return true
		}
	})
	return nil
}
