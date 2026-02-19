package utils

import "strings"

// IsUniqueConstraintError reports SQLite duplicate key violations.
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
