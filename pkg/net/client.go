package net

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/tochemey/goakt/v4/actor"

	"codeberg.org/ApoZero/voxel-server-go/pkg/config"
	"codeberg.org/ApoZero/voxel-server-go/pkg/crypto"
	"codeberg.org/ApoZero/voxel-server-go/pkg/domain"
	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/minecraft"
)

type PacketDecoder interface {
	Decode(data []byte)
}

type RawPacketDecoder struct{}

func (d *RawPacketDecoder) Decode(data []byte) {}

type ConnectionWriter struct {
	from *actor.PID
	to   *actor.PID
}

func (w *ConnectionWriter) Write(data []byte) (int, error) {
	if w.from == nil || w.to == nil {
		return 0, fmt.Errorf("ConnectionWriter PID is not set")
	}

	err := w.from.Tell(context.Background(), w.to, &application.DataExchangeMessage{
		Data: data,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to send client message: %w", err)
	}

	return len(data), nil
}

type ReceivePacket struct{}

type MinecraftClient struct {
	compressPackets bool
	username        string
	verifyToken     []byte
	connWriter      io.Writer
	packetReader    *MinecraftPacketReader
	packetDecoder   PacketDecoder
}

func NewMinecraftClient() *MinecraftClient {
	ret := &MinecraftClient{
		packetReader:  NewMinecraftPacketReader(),
		packetDecoder: &RawPacketDecoder{},
	}
	return ret
}

func (c *MinecraftClient) PreStart(ctx *actor.Context) error {
	return nil
}

func (c *MinecraftClient) Receive(ctx *actor.ReceiveContext) {
	switch ctx.Message().(type) {
	case *actor.PostStart:
		c.connWriter = &ConnectionWriter{
			from: ctx.Self(),
			to:   ctx.Self().Parent(),
		}
		ctx.Become(c.ReceiveHandshake)
	default:
		logrus.Warnf("(MinecraftClient::Receive) Received unknown message type: %T", ctx.Message())
	}
}

func (c *MinecraftClient) PostStop(ctx *actor.Context) error {
	logrus.Infof("Shutting down MinecraftClient actor for: %s", ctx.ActorName())
	return nil
}

func (c *MinecraftClient) tryReceivePacket(ctx *actor.ReceiveContext) {
	if c.packetReader.remainingBufferLength == 0 {
		return
	}

	c.packetReader.currentBuffer = c.packetReader.remainingBuffer
	c.packetReader.currentBufferLength = c.packetReader.remainingBufferLength
	c.packetReader.currentBufferOffset = 0

	packetLength, err := c.packetReader.ReadVarInt()
	if err != nil {
		logrus.Errorf("Failed to read packet length: %s", err.Error())
		return
	}
	if packetLength <= 0 {
		logrus.Warnf("Received packet with non-positive length: %d", packetLength)
		return
	}

	neededLength := packetLength + c.packetReader.currentBufferOffset
	if neededLength > c.packetReader.remainingBufferLength {
		// logrus.Warnf("Received packet with length %d but buffer has %d bytes, waiting for more data", packetLength, c.packetReader.remainingBufferLength)
		return
	}

	err = ctx.Self().Tell(ctx.Context(), ctx.Self(), &ReceivePacket{})
	if err != nil {
		logrus.Errorf("Failed to send ReceivePacket message to self: %s", err.Error())
	}
}

func (c *MinecraftClient) ReceiveHandshake(ctx *actor.ReceiveContext) {
	var err error

	switch msg := ctx.Message().(type) {
	case *application.DataExchangeMessage:
		c.packetDecoder.Decode(msg.Data)
		if err = c.packetReader.AddData(msg.Data); err != nil {
			ctx.Shutdown()
			return
		}

		c.tryReceivePacket(ctx)
	case *ReceivePacket:
		if err = c.packetReader.StartReadingPacket(); err != nil {
			if errors.Is(err, &ReachedPacketEndError{}) {
				logrus.Infof("Reached packet end")
				return
			}

			logrus.Errorf("Error reading next packet: %s", err.Error())
			return
		}

		if c.packetReader.GetPacketId() == 0x00 {
			err = c.onIntention(ctx, c.packetReader)
		}

		if err != nil {
			logrus.Errorf("Error handling handshake packet: %s", err.Error())
			return
		}

		c.tryReceivePacket(ctx)
	default:
		logrus.Warnf("(MinecraftClient::ReceiveHandshake) Received unknown message type: %T", ctx.Message())
	}
}

func (c *MinecraftClient) Write(ctx *actor.ReceiveContext, writer *MinecraftPacketWriter) {
	if c.connWriter == nil {
		logrus.Errorf("Connection writer is not set, cannot send message")
		return
	}

	_, err := c.connWriter.Write(writer.Bytes())
	if err != nil {
		logrus.Errorf("Failed to write to connection: %s", err.Error())
	}
}

func (c *MinecraftClient) ReceiveStatus(ctx *actor.ReceiveContext) {
	var err error

	switch msg := ctx.Message().(type) {
	case *application.DataExchangeMessage:
		c.packetDecoder.Decode(msg.Data)
		if err = c.packetReader.AddData(msg.Data); err != nil {
			ctx.Shutdown()
			return
		}

		c.tryReceivePacket(ctx)
	case *ReceivePacket:
		if err = c.packetReader.StartReadingPacket(); err != nil {
			if errors.Is(err, &ReachedPacketEndError{}) {
				return
			}

			logrus.Errorf("Error reading next packet: %s", err.Error())
			return
		}

		if c.packetReader.GetPacketId() == 0x00 {
			err = c.onVersion(ctx, c.packetReader)
		} else if c.packetReader.GetPacketId() == 0x01 {
			err = c.onPing(ctx, c.packetReader)
		}

		if err != nil {
			logrus.Errorf("Error handling status packet: %s", err.Error())
			return
		}

		c.tryReceivePacket(ctx)
	default:
		logrus.Warnf("(MinecraftClient::ReceiveStatus) Received unknown message type: %T", ctx.Message())
	}
}

func (c *MinecraftClient) ReceiveLogin(ctx *actor.ReceiveContext) {
	var err error

	switch msg := ctx.Message().(type) {
	case *application.DataExchangeMessage:
		c.packetDecoder.Decode(msg.Data)
		if err = c.packetReader.AddData(msg.Data); err != nil {
			ctx.Shutdown()
			return
		}

		c.tryReceivePacket(ctx)
	case *ReceivePacket:
		if err = c.packetReader.StartReadingPacket(); err != nil {
			if errors.Is(err, &ReachedPacketEndError{}) {
				return
			}

			logrus.Errorf("Error starting to read packet: %s", err.Error())
			return
		}

		if c.packetReader.GetPacketId() == 0x00 {
			err = c.onLoginStart(ctx, c.packetReader)
		} else if c.packetReader.GetPacketId() == 0x01 {
			err = c.onEncryptionResponse(ctx, c.packetReader)
		} else if c.packetReader.GetPacketId() == 0x03 {
			ctx.Become(c.ReceiveConfiguration)
		} else {
			logrus.Warnf("Received unknown packet ID during login: %d", c.packetReader.GetPacketId())
			return
		}

		if err != nil {
			logrus.Errorf("Error handling status packet: %s", err.Error())
			return
		}

		c.tryReceivePacket(ctx)
	default:
		logrus.Warnf("(MinecraftClient::ReceiveLogin) Received unknown message type: %T", ctx.Message())
	}
}

func (c *MinecraftClient) ReceiveConfiguration(ctx *actor.ReceiveContext) {
	var err error

	switch msg := ctx.Message().(type) {
	case *application.DataExchangeMessage:
		c.packetDecoder.Decode(msg.Data)
		if err = c.packetReader.AddData(msg.Data); err != nil {
			ctx.Shutdown()
			return
		}

		c.tryReceivePacket(ctx)
	case *ReceivePacket:
		if err = c.packetReader.StartReadingPacket(); err != nil {
			if errors.Is(err, &ReachedPacketEndError{}) {
				return
			}
		}

		logrus.Infof("Received packet with ID %d during configuration phase, but no handler is implemented for this phase", c.packetReader.GetPacketId())

	default:
		logrus.Warnf("(MinecraftClient::ReceiveConfiguration) Received unknown message type: %T", ctx.Message())
	}
}

func (c *MinecraftClient) onIntention(ctx *actor.ReceiveContext, reader *MinecraftPacketReader) error {
	procotolVersion, err := reader.ReadVarInt()
	if err != nil {
		return NewInvalidProtocolVersionError(procotolVersion)
	}

	_, err = reader.ReadString()
	if err != nil {
		logrus.Errorf("Failed to read address: %s", err.Error())
		return NewWrappedHandlerError(err, "failed to read address")
	}

	_, err = reader.ReadUint16()
	if err != nil {
		return NewWrappedHandlerError(err, "failed to read port")
	}

	intent, err := reader.ReadVarInt()
	if err != nil {
		return NewWrappedHandlerError(err, "failed to read intent")
	}

	switch intent {
	case 1:
		logrus.Trace("Client intends to query server status")
		ctx.Become(c.ReceiveStatus)
	case 2:
		logrus.Trace("Client intends to login")
		ctx.Become(c.ReceiveLogin)
	default:
		logrus.Warnf("Received unknown intent: %d", intent)
		return nil
	}

	return nil
}

func (c *MinecraftClient) onVersion(ctx *actor.ReceiveContext, _ *MinecraftPacketReader) error {
	configProvider := ctx.Extension(application.ConfigProviderID).(config.Provider)

	statusResponse := minecraft.StatusResponse{
		Version: minecraft.StatusResponse_Version{
			Name:     "1.21.11",
			Protocol: 774,
		},
		Players: minecraft.StatusResponse_Players{
			Max:    configProvider.GetMaxPlayers(),
			Online: 0,
			Sample: nil,
		},
		Description: minecraft.StatusResponse_Description{
			Text: configProvider.GetMOTD(),
		},
	}

	jsonStatusResponse, err := json.Marshal(statusResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal status response: %w", err)
	}

	writer := NewMinecraftPacketWriter(0x00, c.compressPackets)
	writer.WriteString(string(jsonStatusResponse))
	c.Write(ctx, writer)

	return nil
}

func (c *MinecraftClient) onPing(ctx *actor.ReceiveContext, reader *MinecraftPacketReader) error {
	timestamp, err := reader.ReadLong()
	if err != nil {
		return NewWrappedHandlerError(err, "cannot read timestamp")
	}

	writer := NewMinecraftPacketWriter(0x01, c.compressPackets)
	writer.WriteLong(timestamp)
	c.Write(ctx, writer)

	return nil
}

func (c *MinecraftClient) onLoginStart(ctx *actor.ReceiveContext, reader *MinecraftPacketReader) error {
	configProvider := ctx.Extension(application.ConfigProviderID).(config.Provider)

	username, err := reader.ReadString()
	if err != nil {
		return NewWrappedHandlerError(err, "failed to read username")
	}

	verifyToken, err := crypto.GenerateRandomBytes(16)
	if err != nil {
		return NewWrappedHandlerError(err, "failed to generate verify token")
	}

	c.username = username
	c.verifyToken = verifyToken

	logrus.Tracef("Player %s is trying to log in", username)

	writer := NewMinecraftPacketWriter(0x01, c.compressPackets)
	writer.WriteString("")
	writer.WriteBytes(configProvider.GetPublicKey())
	writer.WriteBytes(c.verifyToken)
	writer.WriteBool(true)
	c.Write(ctx, writer)

	return nil
}

func (c *MinecraftClient) setReadWriteCodecs(sharedSecret []byte) error {
	b, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	c.connWriter = cipher.StreamWriter{
		S: crypto.NewCFB8Encrypter(b, sharedSecret),
		W: c.connWriter,
	}

	c.packetDecoder = crypto.NewCFB8PacketDecoder(
		crypto.NewCFB8Decrypter(b, sharedSecret),
	)

	return nil
}

func (c *MinecraftClient) onEncryptionResponse(ctx *actor.ReceiveContext, reader *MinecraftPacketReader) error {
	configProvider := ctx.Extension(application.ConfigProviderID).(config.Provider)
	gameProfileRepo := ctx.Extension(application.GameProfileRepositoryID).(domain.GameProfileRepository)

	encryptedSharedSecret, err := reader.ReadBytes()
	if err != nil {
		return NewWrappedHandlerError(err, "failed to read encrypted shared secret")
	}

	encryptedVerifyToken, err := reader.ReadBytes()
	if err != nil {
		return NewWrappedHandlerError(err, "failed to read encrypted verify token")
	}

	sharedSecret, err := rsa.DecryptPKCS1v15(nil, configProvider.GetPrivateKey(), encryptedSharedSecret) // nolint:staticcheck
	if err != nil {
		return NewWrappedHandlerError(err, "failed to decrypt shared secret")
	}

	verifyToken, err := rsa.DecryptPKCS1v15(nil, configProvider.GetPrivateKey(), encryptedVerifyToken) // nolint:staticcheck
	if err != nil {
		return NewWrappedHandlerError(err, "failed to decrypt verify token")
	}

	if !bytes.Equal(c.verifyToken, verifyToken) {
		return NewWrappedHandlerError(fmt.Errorf("verify token mismatch"), "verify token mismatch")
	}

	authDigest := crypto.ComputeAuthDigest(sharedSecret, configProvider.GetPublicKey())

	gameProfile, err := gameProfileRepo.GetGameProfile(ctx.Context(), c.username, authDigest)
	if err != nil {
		return fmt.Errorf("failed to get joined data: %w", err)
	}

	if err = c.setReadWriteCodecs(sharedSecret); err != nil {
		return fmt.Errorf("failed to set read/write codecs: %w", err)
	}

	logrus.Tracef("Player %s has logged in with UUID %s", gameProfile.Name, gameProfile.ID)

	writer := NewMinecraftPacketWriter(0x02, c.compressPackets)
	err = writer.WriteUUID16(gameProfile.ID)
	if err != nil {
		return fmt.Errorf("failed to write UUID: %w", err)
	}
	writer.WriteString(gameProfile.Name)
	writer.WriteVarInt(int32(len(gameProfile.Properties)))
	for _, property := range gameProfile.Properties {
		writer.WriteString(property.Name)
		writer.WriteString(property.Value)
		if property.Signature != "" {
			writer.WriteBool(true)
			writer.WriteString(property.Signature)
		} else {
			writer.WriteBool(false)
		}
	}

	c.Write(ctx, writer)

	return nil
}
