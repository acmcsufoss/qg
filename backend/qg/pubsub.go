package qg

import "context"

// CommandHandlerFactory is a factory for creating command handlers. Each
// command handler is considered its own session.
type CommandHandlerFactory interface {
	// NewCommandHandler creates a new command handler. Events will be published
	// to the given event channel.
	NewCommandHandler(ctx context.Context, evs chan<- IEvent) (CommandHandler, error)
}

// CommandHandler describes a command handler. Usually, a game manager will
// implement this interface to manage a game.
type CommandHandler interface {
	// HandleCommand handles a command.
	HandleCommand(ctx context.Context, cmd ICommand) error
	// Close closes the command handler. It should unsubscribe the event channel
	// from all topics.
	Close() error
}
