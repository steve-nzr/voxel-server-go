package net

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/tochemey/goakt/v4/actor"

	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
)

type Server struct {
	options *application.Options_Server
	clients []*actor.PID

	listener net.Listener
	stopChan chan struct{}
	selfPID  *actor.PID
}

func NewServer(options *application.Options_Server) *Server {
	return &Server{
		options:  options,
		stopChan: make(chan struct{}),
		clients:  make([]*actor.PID, 0),
	}
}

func (s *Server) PreStart(ctx *actor.Context) error {
	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", s.options.Address, s.options.Port))
	if err != nil {
		return err
	}

	s.listener = listener
	return nil
}

func (s *Server) Receive(ctx *actor.ReceiveContext) {
	switch ctx.Message().(type) {
	case *actor.PostStart:
		s.selfPID = ctx.Self()
		go s.acceptLoop()
	default:
		logrus.Warnf("Received unknown message type: %T", ctx.Message())
	}
}

func (s *Server) PostStop(ctx *actor.Context) error {
	close(s.stopChan)
	for _, client := range s.clients {
		err := client.Shutdown(ctx.Context())
		if err != nil {
			logrus.Errorf("Error shutting down client actor %s: %s", client.ID(), err.Error())
		}
	}
	return s.listener.Close()
}

func (s *Server) acceptLoop() {
	actorFactory := s.selfPID.ActorSystem().Extension(application.ActorFactoryID).(application.ActorFactory)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				break
			}

			pid, err := actorFactory.CreateConnection(ctx, s.selfPID, conn)
			if err != nil {
				logrus.Errorf("Error spawning child actor: %s", err.Error())
				_ = conn.Close()
				continue
			}

			logrus.Infof("Spawned client actor with PID: %s", pid.ID())
			s.clients = append(s.clients, pid)
		}
	}
}
