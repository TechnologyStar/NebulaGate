package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseRelativeWindow parses a shorthand time window string (e.g. "1h", "24h", "7d")
// and returns the start time in UTC. When the window represents all time, the returned
// time is zero and the boolean flag will be true.
func ParseRelativeWindow(window string) (time.Time, bool, error) {
	normalized := strings.TrimSpace(strings.ToLower(window))
	if normalized == "" || normalized == "all" || normalized == "all_time" || normalized == "alltime" {
		return time.Time{}, true, nil
	}

	now := time.Now().UTC()
	switch normalized {
	case "1h", "1hour", "hour", "1hr":
		return now.Add(-1 * time.Hour), false, nil
	case "24h", "day", "1d", "24hour", "24hours":
		return now.Add(-24 * time.Hour), false, nil
	case "7d", "week", "7day", "7days", "1w":
		return now.Add(-7 * 24 * time.Hour), false, nil
	case "30d", "month", "30day", "30days", "1m":
		return now.Add(-30 * 24 * time.Hour), false, nil
	case "365d", "year", "365day", "365days", "1y":
		return now.Add(-365 * 24 * time.Hour), false, nil
	}

	if len(normalized) > 1 {
		unit := normalized[len(normalized)-1]
		valueStr := normalized[:len(normalized)-1]
		amount, err := strconv.Atoi(valueStr)
		if err == nil {
			switch unit {
			case 'h':
				return now.Add(-time.Duration(amount) * time.Hour), false, nil
			case 'd':
				return now.Add(-time.Duration(amount) * 24 * time.Hour), false, nil
			case 'w':
				return now.Add(-time.Duration(amount) * 7 * 24 * time.Hour), false, nil
			case 'm':
				return now.Add(-time.Duration(amount) * 30 * 24 * time.Hour), false, nil
			case 'y':
				return now.Add(-time.Duration(amount) * 365 * 24 * time.Hour), false, nil
			}
		}
	}

	return time.Time{}, false, fmt.Errorf("unsupported window: %s", window)
}
