package utils

import "time"

// NowUTC returns current timestamp in UTC timezone.
func NowUTC() time.Time {
	return time.Now().UTC()
}
