package net

import (
	"errors"
	"fmt"
)

type HandlerError struct {
	error
}

func NewHandlerError(message string) *HandlerError {
	return &HandlerError{
		error: errors.New(message),
	}
}

func NewWrappedHandlerError(err error, message string) *HandlerError {
	return &HandlerError{
		error: fmt.Errorf("%s: %w", message, err),
	}
}

type InvalidPacketIdError struct {
	PacketID int32
}

func NewInvalidPacketIdError(packetId int32) *InvalidPacketIdError {
	return &InvalidPacketIdError{
		PacketID: packetId,
	}
}

func (e *InvalidPacketIdError) Error() string {
	return fmt.Sprintf("invalid packet ID: %d", e.PacketID)
}

type InvalidProtocolVersionError struct {
	ProtocolVersion int
}

func NewInvalidProtocolVersionError(protocolVersion int) *InvalidProtocolVersionError {
	return &InvalidProtocolVersionError{
		ProtocolVersion: protocolVersion,
	}
}

func (e *InvalidProtocolVersionError) Error() string {
	return fmt.Sprintf("invalid protocol version: %d", e.ProtocolVersion)
}

type ReachedPacketEndError struct{}

func NewReachedPacketEndError() *ReachedPacketEndError {
	return &ReachedPacketEndError{}
}

func (e *ReachedPacketEndError) Error() string {
	return "reached end of packet"
}

type MissingDataError struct{}

func NewMissingDataError() *MissingDataError {
	return &MissingDataError{}
}

func (e *MissingDataError) Error() string {
	return "missing data in packet"
}
