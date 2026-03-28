package crypto

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

func ComputeAuthDigest(sharedSecret []byte, publicKey []byte) string {
	sha1 := sha1.New()
	sha1.Write([]byte(""))
	sha1.Write(sharedSecret)
	sha1.Write(publicKey)
	return computeAuthDigest(sha1.Sum(nil))
}

// computeAuthDigest computes a special SHA-1 digest required for Minecraft web
// authentication on Premium servers (online-mode=true).
// Source: http://wiki.vg/Protocol_Encryption#Server
//
// Also many, many thanks to SirCmpwn and his wonderful gist (C#):
// https://gist.github.com/SirCmpwn/404223052379e82f91e6
func computeAuthDigest(hash []byte) string {
	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}
