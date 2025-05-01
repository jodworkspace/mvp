package utils

import "strconv"

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
