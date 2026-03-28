package net

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/tochemey/goakt/v4/actor"

	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
)

type Conn = net.Conn

type ConnectionActor struct {
	conn       net.Conn
	bufferSize int
	stopChan   chan struct{}
	selfPID    *actor.PID
	clientPID  *actor.PID
}

func NewConnectionActor(conn net.Conn, bufferSize int) *ConnectionActor {
	return &ConnectionActor{
		conn:       conn,
		bufferSize: bufferSize,
		stopChan:   make(chan struct{}),
	}
}

func (c *ConnectionActor) PreStart(ctx *actor.Context) error {
	return nil
}

func (c *ConnectionActor) Receive(ctx *actor.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case *actor.PostStart:
		c.onPostStart(ctx)
	case *application.DataExchangeMessage:
		c.onDataExchangeMessage(msg)
	default:
		logrus.Warnf("Received unknown message type: %T", ctx.Message())
	}
}

func (c *ConnectionActor) PostStop(ctx *actor.Context) error {
	logrus.Infof("Shutting down connection actor for: %s", c.conn.RemoteAddr().String())
	c.closeConnection()
	return nil
}

func (c *ConnectionActor) onPostStart(ctx *actor.ReceiveContext) {
	actorFactory := ctx.Extension(application.ActorFactoryID).(application.ActorFactory)

	clientPID, err := actorFactory.CreateClient(ctx.Context(), ctx.Self())
	if err != nil {
		logrus.Errorf("Failed to create client actor for connection: %s", c.conn.RemoteAddr().String())
		ctx.Stop(ctx.Self())
		return
	}

	c.selfPID = ctx.Self()
	c.clientPID = clientPID

	go c.receiveLoop()
}

func (c *ConnectionActor) onDataExchangeMessage(msg *application.DataExchangeMessage) {
	_, err := c.conn.Write(msg.Data)
	if err != nil {
		logrus.Errorf("Error writing to connection: %s", err.Error())
	}
}

func (c *ConnectionActor) closeConnection() {
	logrus.Infof("Closing connection: %s", c.conn.RemoteAddr().String())
	err := c.conn.Close()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		logrus.Errorf("Error closing connection: %s", err.Error())
	}
}

func (c *ConnectionActor) receiveLoop() {
	buf := make([]byte, c.bufferSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logrus.Infof("Started receive loop for connection: %s", c.conn.RemoteAddr().String())

	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				break
			}

			logrus.Errorf("Error reading from connection: %s", err.Error())
			break
		}

		chunk := make([]byte, n)
		copy(chunk, buf[:n])

		err = c.selfPID.Tell(ctx, c.clientPID, &application.DataExchangeMessage{Data: chunk})
		if err != nil {
			logrus.Errorf("Error sending message to client actor: %s", err.Error())
			continue
		}
	}

	logrus.Infof("Exiting receive loop for connection: %s", c.conn.RemoteAddr().String())

	err := c.selfPID.Shutdown(context.Background())
	if err != nil {
		logrus.Errorf("Error shutting down connection actor: %s", err.Error())
	}
}
