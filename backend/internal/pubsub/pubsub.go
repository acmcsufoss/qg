package pubsub

import (
	"context"
	"sync"

	"oss.acmcsuf.com/qg/backend/qg"
)

// Publisher implements an event publisher.
type Publisher struct {
	subs sync.Map
}

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

// CopyTo copies all event channels from the publisher to the given publisher.
func (p *Publisher) CopyTo(other *Publisher) {
	p.subs.Range(func(key, _ any) bool {
		other.subs.Store(key, struct{}{})
		return true
	})
}
