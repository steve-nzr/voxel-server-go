package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tochemey/goakt/v4/actor"

	"codeberg.org/ApoZero/voxel-server-go/pkg/application"
	"codeberg.org/ApoZero/voxel-server-go/pkg/config"
	appdomain "codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
	"codeberg.org/ApoZero/voxel-server-go/pkg/infra/api/mojang"
	"codeberg.org/ApoZero/voxel-server-go/pkg/infra/repo"
)

const (
	MinecraftPrivateKeyBits = 1024
)

func Launch(ctx context.Context, opts appdomain.Options) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, MinecraftPrivateKeyBits)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	mojangAPI := mojang.NewMojangAPI(&opts.Mojang)
	configProvider := config.NewStaticConfigProvider(privateKey)
	gameProfileRepo := repo.NewGameProfileRepository_Mojang(mojangAPI)

	actorFactory := application.NewActorFactory(&opts)

	sys, err := actor.NewActorSystem(
		"voxel-server-go",
		actor.WithExtensions(actorFactory, configProvider, gameProfileRepo),
	)
	if err != nil {
		return fmt.Errorf("failed to create actor system: %w", err)
	}

	go func() {
		for !sys.Running() {
			logrus.Info("Waiting for actor system to start...")
			time.Sleep(time.Millisecond)
		}

		pid := actorFactory.CreateServer(ctx, sys)
		logrus.Infof("Server started with PID: %s", pid.ID())
	}()

	sys.Run(ctx, func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil })
	return nil
}
