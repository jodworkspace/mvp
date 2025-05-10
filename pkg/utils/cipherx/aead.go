package cipherx

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type AEAD struct {
	cipher.AEAD
}

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

func (a *AEAD) Encrypt(plaintext []byte, data ...[]byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return []byte(""), nil
	}

	nonce := make([]byte, a.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := a.Seal(nil, nonce, plaintext, func() []byte {
		if len(data) > 0 {
			return data[0]
		}
		return nil
	}())

	return append(nonce, ciphertext...), nil
}

func (a *AEAD) Decrypt(ciphertextWithNonce []byte, data ...[]byte) ([]byte, error) {
	if len(ciphertextWithNonce) == 0 {
		return []byte(""), nil
	}

	nonce := ciphertextWithNonce[:a.NonceSize()]
	ciphertext := ciphertextWithNonce[a.NonceSize():]

	plaintext, err := a.Open(nil, nonce, ciphertext, func() []byte {
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
