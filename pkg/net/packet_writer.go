package net

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unsafe"

	"codeberg.org/ApoZero/voxel-server-go/pkg/crypto"
)

type MinecraftPacketWriter struct {
	packetId   int32
	compressed bool
	buffer     []byte
}

func NewMinecraftPacketWriter(packetId int32, compressed bool) *MinecraftPacketWriter {
	return &MinecraftPacketWriter{
		packetId:   packetId,
		compressed: compressed,
		buffer:     make([]byte, 0, 64),
	}
}

func (w *MinecraftPacketWriter) WriteString(value string) {
	length := crypto.ComputeVarInt(int32(len(value)))
	w.buffer = append(w.buffer, length...)
	w.buffer = append(w.buffer, []byte(value)...)
}

func (w *MinecraftPacketWriter) WriteBytes(value []byte) {
	length := crypto.ComputeVarInt(int32(len(value)))
	w.buffer = append(w.buffer, length...)
	w.buffer = append(w.buffer, value...)
}

func (w *MinecraftPacketWriter) WriteFixedLengthBytes(value []byte) {
	w.buffer = append(w.buffer, value...)
}

func (w *MinecraftPacketWriter) WriteBool(value bool) {
	if value {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}
}

func (w *MinecraftPacketWriter) WriteByte(value byte) {
	w.buffer = append(w.buffer, value)
}

func (w *MinecraftPacketWriter) WriteInt16(value int16) {
	buffer := make([]byte, unsafe.Sizeof(value))
	binary.BigEndian.PutUint16(buffer, uint16(value))
	w.buffer = append(w.buffer, buffer...)
}

func (w *MinecraftPacketWriter) WriteInt32(value int32) {
	buffer := make([]byte, unsafe.Sizeof(value))
	binary.BigEndian.PutUint32(buffer, uint32(value))
	w.buffer = append(w.buffer, buffer...)
}

func (w *MinecraftPacketWriter) WriteInt64(value int64) {
	buffer := make([]byte, unsafe.Sizeof(value))
	binary.BigEndian.PutUint64(buffer, uint64(value))
	w.buffer = append(w.buffer, buffer...)
}

func (w *MinecraftPacketWriter) WriteVarInt(value int32) {
	varInt := crypto.ComputeVarInt(value)
	w.buffer = append(w.buffer, varInt...)
}

func (w *MinecraftPacketWriter) WriteLong(value int64) {
	w.WriteInt64(value)
}

func (w *MinecraftPacketWriter) WriteUUID16(id string) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return fmt.Errorf("invalid uuid hex: %w", err)
	}
	if len(b) != 16 {
		return fmt.Errorf("invalid uuid length: got %d, want 16", len(b))
	}
	w.WriteFixedLengthBytes(b)
	return nil
}

func (w MinecraftPacketWriter) Bytes() []byte {
	packetIdVarInt := crypto.ComputeVarInt(w.packetId)
	packetLength := crypto.ComputeVarInt(int32(len(packetIdVarInt) + len(w.buffer)))

	buffer := make([]byte, 0, len(packetIdVarInt)+len(w.buffer)+len(packetLength))
	buffer = append(buffer, packetLength...)
	buffer = append(buffer, packetIdVarInt...)
	buffer = append(buffer, w.buffer...)

	return buffer
}
