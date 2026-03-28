package crypto

import "fmt"

type CryptographyError struct {
	error
}

func NewCryptographyError(err error, message string) *CryptographyError {
	return &CryptographyError{
		error: fmt.Errorf("%s: %w", message, err),
	}
}
