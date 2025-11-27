package util

import (
	"fmt"
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

// FormatCheckpointDate formats a datetime for checkpoint display
// Format: "Dec 25, 2025, 3:00 am" or "Dec 25, 3 am" (if current year and minutes are 00)
func FormatCheckpointDate(datetime time.Time) string {
	now := time.Now()
	currentYear := now.Year()
	year := datetime.Year()
	month := datetime.Month()
	day := datetime.Day()
	hour := datetime.Hour()
	minute := datetime.Minute()

	// Format month abbreviation (e.g., "Dec")
	monthAbbr := month.String()[:3]

	// Format time in 12-hour format
	ampm := "am"
	displayHour := hour
	if hour == 0 {
		displayHour = 12
	} else if hour == 12 {
		ampm = "pm"
	} else if hour > 12 {
		displayHour = hour - 12
		ampm = "pm"
	}

	// Build the date part
	datePart := fmt.Sprintf("%s %d", monthAbbr, day)

	// Add year if not current year
	if year != currentYear {
		datePart += fmt.Sprintf(", %d", year)
	}

	// Format time part
	var timePart string
	if minute == 0 {
		timePart = fmt.Sprintf("%d%s", displayHour, ampm)
	} else {
		timePart = fmt.Sprintf("%d:%02d %s", displayHour, minute, ampm)
	}

	return fmt.Sprintf("%s, %s", datePart, timePart)
}

// FormatCountdown formats the time until a datetime as a countdown string
// Returns: "In 3 days", "In 5 hours 30 minutes", "In 45 minutes", etc.
func FormatCountdown(datetime time.Time) string {
	now := time.Now()
	duration := datetime.Sub(now)

	if duration < 0 {
		return "Past"
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if days == 1 {
			return "In 1 day"
		}
		return fmt.Sprintf("In %d days", days)
	}

	if hours > 0 {
		if minutes > 0 {
			if hours == 1 {
				if minutes == 1 {
					return "In 1 hour 1 minute"
				}
				return fmt.Sprintf("In 1 hour %d minutes", minutes)
			}
			if minutes == 1 {
				return fmt.Sprintf("In %d hours 1 minute", hours)
			}
			return fmt.Sprintf("In %d hours %d minutes", hours, minutes)
		}
		if hours == 1 {
			return "In 1 hour"
		}
		return fmt.Sprintf("In %d hours", hours)
	}

	if minutes > 0 {
		if minutes == 1 {
			return "In 1 minute"
		}
		return fmt.Sprintf("In %d minutes", minutes)
	}

	return "Now"
}
