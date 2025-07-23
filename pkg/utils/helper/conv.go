package helper

import (
	"encoding/base64"
	"strconv"
)

func Base64Encode(s []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
	base64.StdEncoding.Encode(encoded, s)

	return encoded
}

func Base64Decode(s []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(s)))

	n, err := base64.StdEncoding.Decode(decoded, s)
	if err != nil {
		return nil, err
	}

	return decoded[:n], nil
}

func StringToUint64(s string) uint64 {
	if s == "" {
		return 0
	}

	number, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}

	return number
}
