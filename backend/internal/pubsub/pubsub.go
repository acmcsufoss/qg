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

type subscribable struct {
	publisher *Publisher
	channel   chan<- qg.IEvent
}

// Subscribe implements the Subscriber interface.
func (p *Publisher) Subscribe(out chan<- qg.IEvent) {
	p.subs.Store(subscribable{channel: out}, struct{}{})
}

// Unsubscribe implements the Subscriber interface.
func (p *Publisher) Unsubscribe(out chan<- qg.IEvent) {
	p.subs.Delete(subscribable{channel: out})
}

// SubscribePublisher subscribes the given publisher to the publisher.
func (p *Publisher) SubscribePublisher(given *Publisher) {
	p.subs.Store(subscribable{publisher: given}, struct{}{})
}

// UnsubscribePublisher unsubscribes the given publisher from the publisher.
func (p *Publisher) UnsubscribePublisher(given *Publisher) {
	p.subs.Delete(subscribable{publisher: given})
}

// Publish implements the Publisher interface.
func (p *Publisher) Publish(ctx context.Context, msg qg.IEvent) {
	p.subs.Range(func(k, _ any) bool {
		s := k.(subscribable)

		switch {
		case s.publisher != nil:
			s.publisher.Publish(ctx, msg)
		case s.channel != nil:
			select {
			case <-ctx.Done():
				return false
			case s.channel <- msg:
				// ok
			default:
				p.subs.Delete(k)
			}
		}

		return true
	})
}
