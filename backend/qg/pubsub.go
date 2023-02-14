package qg

import "context"

// Subscriber describes a subscriber.
type Subscriber interface {
	// Subscribe subscribes to the given topics. The given channel will
	// receive messages published to the topics. If no topics are given, the
	// channel will receive all messages. The returned channel will be closed
	// when the context is canceled.
	Subscribe(ctx context.Context, out chan<- Event) error
	// Unsubscribe unsubscribes from the given topics. If no topics are given,
	// the given cnannel will be unsubscribed from all topics.
	Unsubscribe(ctx context.Context, forCh chan<- Event) error
}

// Publisher describes a publisher.
type Publisher interface {
	Subscriber
	// Publish publishes the given message to all subscribers.
	Publish(ctx context.Context, msg Event) error
}

// CommandHandlerFactory is a factory for creating command handlers. Each
// command handler is considered its own session.
type CommandHandlerFactory interface {
	CommandHandler() CommandHandler
}

// CommandHandler describes a command handler. Usually, a game manager will
// implement this interface to manage a game.
type CommandHandler interface {
	Subscriber
	// HandleCommand handles a command.
	HandleCommand(ctx context.Context, cmd Command) error
}
