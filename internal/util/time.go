package util

import (
	"strings"
	"time"
)

// ParseDate parses a date string in YYYY-MM-DD format and returns a time.Time at midnight
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ParseTime parses a time string in formats HH:MM or H:MM AM/PM and returns hour and minute
// Ignores seconds if present in 24-hour format
func ParseTime(timeStr string) (hour int, minute int, err error) {
	timeStrLower := strings.ToLower(strings.TrimSpace(timeStr))

	// Extract and remove am/pm if present
	isPM := strings.Contains(timeStrLower, "pm")
	isAM := strings.Contains(timeStrLower, "am")
	if isPM {
		timeStrLower = strings.ReplaceAll(timeStrLower, "pm", "")
	} else if isAM {
		timeStrLower = strings.ReplaceAll(timeStrLower, "am", "")
	}
	timeStrLower = strings.TrimSpace(timeStrLower)

	// If no colon present and am/pm was detected, assume :00 (e.g., "2pm" -> "2:00")
	if (isPM || isAM) && !strings.Contains(timeStrLower, ":") {
		timeStrLower = timeStrLower + ":00"
	}

	// Remove seconds if present (HH:MM:SS -> HH:MM)
	parts := strings.Split(timeStrLower, ":")
	if len(parts) > 2 {
		timeStrLower = parts[0] + ":" + parts[1]
	}

	// Parse HH:MM or H:MM format
	parsedTime, err := time.Parse("15:04", timeStrLower)
	if err != nil {
		parsedTime, err = time.Parse("3:04", timeStrLower)
		if err != nil {
			return 0, 0, err
		}
	}

	hour = parsedTime.Hour()
	minute = parsedTime.Minute()

	// Convert 12-hour format to 24-hour format
	if isPM && hour != 12 {
		hour += 12
	} else if isAM && hour == 12 {
		hour = 0
	}

	return hour, minute, nil
}

func FormatNaturalDate(datetime time.Time) string {
	return datetime.Format(time.DateTime)
}
