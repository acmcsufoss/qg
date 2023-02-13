package qg

import "context"

// TopicEventForGamePlayer returns a subscribable topic that subscribes to the
// topic for all events that are emitted to this player for this game.
func TopicEventForGamePlayer(id GameID, player PlayerName) SubscribableTopic[Event] {
	return SubscribableTopic[Event]{"qg", "event", id, "player", player}
}

func TopicEventForGame(id GameID) SubscribableTopic[Event] {
	return SubscribableTopic[Event]{"qg", "event", id}
}

// SubscribableTopic is a type-safe wrapper for a topic that can be subscribed
// to.
type SubscribableTopic[T any] []string

// Subscribe subscribes a channel to this. The returned channel will receive
// messages published to the topic.
func (t SubscribableTopic[T]) Subscribe(ctx context.Context, subber Subscriber, out chan<- T) error {
	anyOut := make(chan any, len(t))
	if err := subber.Subscribe(ctx, anyOut, t...); err != nil {
		return err
	}
	go func() {
		for msg := range anyOut {
			select {
			case out <- msg.(T):
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// Unsubscribe unsubscribes a channel from this topic.
func (t SubscribableTopic[T]) Unsubscribe(ctx context.Context, subber Subscriber, forCh chan<- any) error {
	if err := subber.Unsubscribe(ctx, forCh, t...); err != nil {
		return err
	}
	return nil
}

// Broadcast broadcasts a message to this topic.
func (t SubscribableTopic[T]) Broadcast(ctx context.Context, pubber Broadcaster, v T) error {
	if err := pubber.Broadcast(ctx, v, t...); err != nil {
		return err
	}
	return nil
}

// PubSub implements a pubsub system.
type PubSub interface {
	Broadcaster
	Subscriber
}

// Subscriber describes a subscriber.
type Subscriber interface {
	// Subscribe subscribes to the given topics. The given channel will
	// receive messages published to the topics. If no topics are given, the
	// channel will receive all messages. The returned channel will be closed
	// when the context is canceled.
	Subscribe(ctx context.Context, out chan<- any, topis ...string) error
	// Unsubscribe unsubscribes from the given topics. If no topics are given,
	// the given cnannel will be unsubscribed from all topics.
	Unsubscribe(ctx context.Context, forCh chan<- any, topics ...string) error
}

// Broadcaster describes a broadcaster.
type Broadcaster interface {
	// Broadcast broadcasts the given message to the given topics. If no topics
	// are given, the message is broadcasted to all topics.
	Broadcast(ctx context.Context, msg any, topics ...string) error
}
