package net

type PacketReader struct {
	Buffer []byte
	Offset int
	Length int
}

func NewPacketReader(buffer []byte, length int) *PacketReader {
	return &PacketReader{
		Buffer: buffer,
		Length: length,
	}
}

func (p *PacketReader) SetData(buf []byte, length int) {
	p.Buffer = buf
	p.Length = length
	p.Offset = 0
}
