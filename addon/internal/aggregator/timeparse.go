package aggregator

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseRouterOSTimestamp(value string, now time.Time) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		v := ts.UTC()
		return &v
	}
	if d, err := parseRouterOSDuration(value); err == nil {
		v := now.UTC().Add(-d)
		return &v
	}
	return nil
}

func parseRouterOSDuration(value string) (time.Duration, error) {
	if strings.Contains(value, ":") {
		parts := strings.Split(value, ":")
		if len(parts) != 3 {
			return 0, fmt.Errorf("invalid hh:mm:ss: %s", value)
		}
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		s, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, err
		}
		return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second, nil
	}

	mult := map[byte]time.Duration{
		'w': 7 * 24 * time.Hour,
		'd': 24 * time.Hour,
		'h': time.Hour,
		'm': time.Minute,
		's': time.Second,
	}

	var dur time.Duration
	number := ""
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			number += string(ch)
			continue
		}
		unit, ok := mult[ch]
		if !ok || number == "" {
			return 0, fmt.Errorf("invalid duration segment: %s", value)
		}
		v, err := strconv.Atoi(number)
		if err != nil {
			return 0, err
		}
		dur += time.Duration(v) * unit
		number = ""
	}
	if number != "" {
		v, err := strconv.Atoi(number)
		if err != nil {
			return 0, err
		}
		dur += time.Duration(v) * time.Second
	}
	if dur == 0 {
		return 0, fmt.Errorf("zero duration")
	}
	return dur, nil
}
