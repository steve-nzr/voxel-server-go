package config

import "crypto/rsa"

type Provider interface {
	GetMOTD() string
	GetMaxPlayers() int
	GetPrivateKey() *rsa.PrivateKey
	GetPublicKey() []byte
}
