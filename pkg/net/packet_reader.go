package net

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const BufferSizeLimit = 1024 * 1024 * 10 // 10 MB

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Packet_format
type MinecraftPacketReader struct {
	remainingBuffer       []byte
	remainingBufferLength int

	currentPacketLength int

	currentBuffer       []byte
	currentBufferLength int
	currentBufferOffset int
	currentPacketID     int
}

func NewMinecraftPacketReader() *MinecraftPacketReader {
	return &MinecraftPacketReader{
		remainingBuffer: make([]byte, 0, 512),
	}
}

func (p *MinecraftPacketReader) ReadPackentLength() (int, error) {
	return p.ReadVarInt()
}

func (p *MinecraftPacketReader) ReadVarInt() (int, error) {
	var result int
	var shift uint32

	for {
		b, err := p.ReadByte()
		if errors.Is(err, io.EOF) {
			return 0, err // EOF
		}

		result |= int(b&0x7F) << shift

		if (b & 0x80) == 0 {
			break
		}

		shift += 7

		if shift >= 32 {
			return 0, fmt.Errorf("varint is too big")
		}
	}

	return result, nil
}

func (m *MinecraftPacketReader) ReadString() (string, error) {
	data, err := m.ReadBytes()
	if err != nil {
		return "", fmt.Errorf("failed to read string bytes: %w", err)
	}

	return string(data), nil
}

func (m *MinecraftPacketReader) ReadBytes() ([]byte, error) {
	length, err := m.ReadVarInt()
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes length: %w", err)
	}

	if length < 0 {
		return nil, fmt.Errorf("invalid bytes length: %d", length)
	}

	buf := make([]byte, length)
	for i := range buf {
		b, err := m.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read bytes byte: %w", err)
		}
		buf[i] = b
	}

	return buf, nil
}

func (p *MinecraftPacketReader) ReadLong() (int64, error) {
	if !p.CanRead(8) {
		return 0, io.ErrUnexpectedEOF
	}

	val := int64(binary.BigEndian.Uint64(p.currentBuffer[p.currentBufferOffset : p.currentBufferOffset+8]))
	p.currentBufferOffset += 8
	return val, nil
}

func (p *MinecraftPacketReader) ReadUint16() (uint16, error) {
	if !p.CanRead(2) {
		return 0, io.ErrUnexpectedEOF
	}

	val := binary.BigEndian.Uint16(p.currentBuffer[p.currentBufferOffset : p.currentBufferOffset+2])
	p.currentBufferOffset += 2
	return val, nil
}

func (p *MinecraftPacketReader) ReadByte() (byte, error) {
	if !p.CanRead(1) {
		return 0, io.EOF
	}

	b := p.currentBuffer[p.currentBufferOffset]
	p.currentBufferOffset++
	return b, nil
}

func (m MinecraftPacketReader) GetPacketId() int {
	return m.currentPacketID
}

func (m *MinecraftPacketReader) AddData(data []byte) error {
	if m.remainingBufferLength >= BufferSizeLimit {
		m.remainingBuffer = nil
		m.remainingBufferLength = 0
		return fmt.Errorf("remaining buffer is too large")
	}

	m.remainingBuffer = append(m.remainingBuffer, data...)
	m.remainingBufferLength += len(data)
	return nil
}

func (m *MinecraftPacketReader) StartReadingPacket() error {
	if !m.HasMoreData() {
		return NewReachedPacketEndError()
	}

	if m.remainingBufferLength < 2 {
		return NewMissingDataError()
	}

	// Don't copy here, just set the current buffer to the remaining buffer and use offsets to track where we are
	m.currentBuffer = m.remainingBuffer
	m.currentBufferLength = m.remainingBufferLength
	m.currentBufferOffset = 0

	packetLength, err := m.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read packet length: %w", err)
	}

	m.currentPacketLength = int(packetLength)

	if m.currentPacketLength > m.remainingBufferLength-m.currentBufferOffset {
		return NewMissingDataError()
	}

	packetEnd := m.currentBufferOffset + m.currentPacketLength

	m.currentBuffer = make([]byte, m.currentPacketLength)
	copy(m.currentBuffer, m.remainingBuffer[m.currentBufferOffset:packetEnd])

	m.currentBufferLength = m.currentPacketLength
	m.currentBufferOffset = 0

	if packetEnd >= m.remainingBufferLength {
		m.remainingBuffer = []byte{}
		m.remainingBufferLength = 0
	} else {
		m.remainingBuffer = m.remainingBuffer[packetEnd:]
		m.remainingBufferLength = len(m.remainingBuffer)
	}

	packetId, err := m.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read packet ID: %w", err)
	}

	m.currentPacketID = packetId
	return nil
}

func (m MinecraftPacketReader) HasMoreData() bool {
	return m.remainingBufferLength > 0
}

func (m MinecraftPacketReader) CanRead(n int) bool {
	return m.currentBufferOffset+n <= m.currentBufferLength
}
