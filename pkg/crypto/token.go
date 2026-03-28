package crypto

import "crypto/rand"

func GenerateRandomBytes(length int) ([]byte, error) {
	randomBytes := make([]byte, length)

	count, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	if count != length {
		return nil, NewCryptographyError(err, "failed to generate random bytes")
	}

	return randomBytes, nil
}
