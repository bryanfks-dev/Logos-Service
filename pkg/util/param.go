package util

import (
	"errors"
	"net/url"
	"time"
)

func ParseDatetimeParam(param string) (time.Time, error) {
	unescaped, err := url.PathUnescape(param)
	if err != nil {
		return time.Time{}, err
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, unescaped); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("invalid datetime param format")
}
