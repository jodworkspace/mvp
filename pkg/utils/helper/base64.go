package helper

import (
	"encoding/base64"
)

func StdBase64Encode(s []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
	base64.StdEncoding.Encode(encoded, s)

	return encoded
}

func StdBase64Decode(s []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(s)))

	n, err := base64.StdEncoding.Decode(decoded, s)
	if err != nil {
		return nil, err
	}

	return decoded[:n], nil
}
