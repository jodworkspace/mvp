package cipherx

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type AEAD struct {
	cipher.AEAD
}

// NewAEAD creates a new AEAD instance in GCM mode using the provided key.
func NewAEAD(key []byte) (*AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AEAD{aead}, nil
}

func MustNewAEAD(key []byte) *AEAD {
	aead, err := NewAEAD(key)
	if err != nil {
		panic(err)
	}

	return aead
}

// Encrypt encrypts the plaintext and returns the ciphertext in Standard Base64 format.
func (a *AEAD) Encrypt(plaintext []byte, data ...[]byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return []byte(""), nil
	}

	nonce := make([]byte, a.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := a.Seal(nil, nonce, plaintext, func() []byte {
		if len(data) > 0 {
			return data[0]
		}
		return nil
	}())

	ciphertext := append(nonce, encrypted...)
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	return []byte(encodedCiphertext), nil
}

// Decrypt decrypts the Standard Base64 encoded ciphertext and returns the plaintext.
func (a *AEAD) Decrypt(encodedCiphertext []byte, data ...[]byte) ([]byte, error) {
	if len(encodedCiphertext) == 0 {
		return []byte(""), nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(encodedCiphertext))
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:a.NonceSize()]
	encrypted := ciphertext[a.NonceSize():]

	plaintext, err := a.Open(nil, nonce, encrypted, func() []byte {
		if len(data) > 0 {
			return data[0]
		}
		return nil
	}())
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
