package util

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"uuid"
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

func ParseUUIDParam(param string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(param)
	if trimmed == "" {
		return uuid.Nil(), errors.New("uuid param is required")
	}

	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.Nil(), errors.New("invalid uuid param format")
	}

	if parsed == uuid.Nil() {
		return uuid.Nil(), errors.New("uuid param must not be nil")
	}

	return parsed, nil
}
