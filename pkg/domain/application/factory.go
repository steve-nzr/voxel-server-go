package application

import (
	"context"
	"net"

	"github.com/tochemey/goakt/v4/actor"
)

type ActorFactory interface {
	// CreateServer doesn't return an error because if it fails, the application should just exit.
	CreateServer(ctx context.Context, sys actor.ActorSystem) *actor.PID
	CreateConnection(ctx context.Context, server *actor.PID, conn net.Conn) (*actor.PID, error)
	CreateClient(ctx context.Context, connection *actor.PID) (*actor.PID, error)
}
