package model

import "time"

// RouterConfig represents a normalized integration configuration payload.
type RouterConfig struct {
	Version         int64     `json:"version"`
	UpdatedAt       time.Time `json:"updated_at"`
	Host            string    `json:"host"`
	Username        string    `json:"username"`
	Password        string    `json:"password"`
	SSL             bool      `json:"ssl"`
	VerifyTLS       bool      `json:"verify_tls"`
	PollIntervalSec int       `json:"poll_interval_sec"`
	Roles           []string  `json:"roles"`
}

func (c RouterConfig) PollInterval() time.Duration {
	interval := time.Duration(c.PollIntervalSec) * time.Second
	if interval < 5*time.Second {
		return 5 * time.Second
	}
	return interval
}

func (c RouterConfig) BaseURL() string {
	scheme := "https"
	if !c.SSL {
		scheme = "http"
	}
	return scheme + "://" + c.Host + "/rest"
}
