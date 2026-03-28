package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"codeberg.org/ApoZero/voxel-server-go/cmd/server"
	appdomain "codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := appdomain.Options{
		Server: appdomain.Options_Server{
			Address:        "127.0.0.1",
			Port:           25565,
			ConnBufferSize: 128,
		},
		Mojang: appdomain.Options_Mojang{
			SessionServerURL: "https://sessionserver.mojang.com/session/minecraft",
		},
	}

	// logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
	})
	logrus.SetLevel(logrus.TraceLevel)

	err := server.Launch(ctx, opts)
	if err != nil {
		logrus.Fatal("Failed to launch server:", err)
	}
}
