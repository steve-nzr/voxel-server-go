package config

import (
	"crypto/rsa"
	"crypto/x509"

	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
)

type StaticConfigProvider struct {
	application.ConfigProviderExtension
	MOTD           string
	MaxPlayers     int
	PrivateKey     *rsa.PrivateKey
	PublicKeyPKCS1 []byte
}

func NewStaticConfigProvider(privateKey *rsa.PrivateKey) *StaticConfigProvider {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	return &StaticConfigProvider{
		MOTD:           "Powered by Golang!",
		MaxPlayers:     100,
		PrivateKey:     privateKey,
		PublicKeyPKCS1: publicKeyBytes,
	}
}

func (p StaticConfigProvider) GetMaxPlayers() int {
	return p.MaxPlayers
}

func (p StaticConfigProvider) GetMOTD() string {
	return p.MOTD
}

func (p StaticConfigProvider) GetPrivateKey() *rsa.PrivateKey {
	return p.PrivateKey
}

func (p StaticConfigProvider) GetPublicKey() []byte {
	return p.PublicKeyPKCS1
}
