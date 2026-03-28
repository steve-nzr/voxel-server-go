package application

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tochemey/goakt/v4/actor"

	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
	"codeberg.org/ApoZero/voxel-server-go/pkg/net"
)

type ActorsFactoryImpl struct {
	application.ActorFactoryExtension
	opts *application.Options
}

func NewActorFactory(opts *application.Options) *ActorsFactoryImpl {
	return &ActorsFactoryImpl{
		opts: opts,
	}
}

func (f *ActorsFactoryImpl) CreateServer(ctx context.Context, sys actor.ActorSystem) *actor.PID {
	pid, err := sys.Spawn(ctx, "server", net.NewServer(&f.opts.Server), actor.WithLongLived())
	if err != nil {
		logrus.Errorf("Failed to spawn server actor: %s", err)
		os.Exit(1)
	}

	return pid
}

func (f *ActorsFactoryImpl) CreateConnection(ctx context.Context, server *actor.PID, conn net.Conn) (*actor.PID, error) {
	pid, err := server.SpawnChild(ctx, uuid.NewString(), net.NewConnectionActor(conn, f.opts.Server.ConnBufferSize), actor.WithLongLived())
	if err != nil {
		return nil, fmt.Errorf("failed to spawn child actor: %w", err)
	}

	return pid, nil
}

func (f *ActorsFactoryImpl) CreateClient(ctx context.Context, connection *actor.PID) (*actor.PID, error) {
	clientPID, err := connection.SpawnChild(ctx, "handler", net.NewMinecraftClient(), actor.WithLongLived())
	if err != nil {
		return nil, fmt.Errorf("failed to create client actor: %w", err)
	}

	return clientPID, nil
}
