package helper

import "time"

func ParseISO8601Date(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
