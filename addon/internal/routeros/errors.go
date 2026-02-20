package routeros

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	goros "github.com/go-routeros/routeros/v3"
)

// ValidationError describes a user-supplied invalid value.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "validation error"
	}
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

// ReconnectError indicates a reconnect loop exceeded retry budget.
type ReconnectError struct {
	Address  string
	Attempts int
	Err      error
}

func (e *ReconnectError) Error() string {
	if e == nil {
		return "reconnect failed"
	}
	return fmt.Sprintf("routeros reconnect to %s failed after %d attempts: %v", e.Address, e.Attempts, e.Err)
}

func (e *ReconnectError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// RuleNotFoundError means firewall rule lookup failed.
type RuleNotFoundError struct {
	ID string
}

func (e *RuleNotFoundError) Error() string {
	if e == nil {
		return "firewall rule not found"
	}
	return fmt.Sprintf("firewall rule %q not found", e.ID)
}

// AddressListNotFoundError means requested address-list entry was not found.
type AddressListNotFoundError struct {
	List    string
	Address string
}

func (e *AddressListNotFoundError) Error() string {
	if e == nil {
		return "address-list entry not found"
	}
	return fmt.Sprintf("address-list entry %q/%q not found", e.List, e.Address)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}

	var nerr net.Error
	if errors.As(err, &nerr) {
		return true
	}

	var deviceErr *goros.DeviceError
	if errors.As(err, &deviceErr) {
		return false
	}

	message := strings.ToLower(err.Error())
	if strings.Contains(message, "broken pipe") {
		return true
	}
	if strings.Contains(message, "connection reset") {
		return true
	}
	if strings.Contains(message, "use of closed network connection") {
		return true
	}
	if strings.Contains(message, "connection refused") {
		return true
	}
	if strings.Contains(message, "i/o timeout") {
		return true
	}
	if strings.Contains(message, "timeout") {
		return true
	}
	return false
}

func isMissingCommandError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such command") ||
		strings.Contains(text, "bad command name") ||
		strings.Contains(text, "input does not match")
}

func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "already have") || strings.Contains(text, "already exists")
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such item") ||
		strings.Contains(text, "not found") ||
		strings.Contains(text, "invalid internal item")
}
