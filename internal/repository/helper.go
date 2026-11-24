package repository

import "time"

func parseSinceTime(since int64) time.Time {
	if since <= 0 {
		return time.Unix(0, 0)
	}

	if since > 1e11 {
		return time.UnixMilli(since)
	}

	return time.Unix(since, 0)
}
