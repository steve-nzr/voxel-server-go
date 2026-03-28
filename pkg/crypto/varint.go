package crypto

func ComputeVarInt(value int32) []byte {
	var buffer []byte
	for {
		temp := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			temp |= 0x80
		}
		buffer = append(buffer, temp)
		if value == 0 {
			break
		}
	}
	return buffer
}
