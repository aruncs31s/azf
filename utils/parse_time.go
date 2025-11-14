package utils

import (
	"time"
)

func ParseTime(someTime string) time.Time {
	format := "2006-01-02"
	if date, err := time.Parse(format, someTime); err == nil {
		return date
	}
	return time.Time{}
}

// ParseDate returns date string in "YYYY-MM-DD" format from an ISO 8601 time input.
func ParseDate(iso8601time any) string {
	if iso8601time == nil {
		return "1970-01-01"
	}
	if t, ok := iso8601time.(time.Time); ok {
		date := t.Format("2006-01-02")
		if len(date) > 10 {
			return date[:10]
		}
		return date
	}
	if t, ok := iso8601time.(*time.Time); ok {
		if t == nil {
			return "1970-01-01"
		}
		date := t.Format("2006-01-02")
		if len(date) > 10 {
			return date[:10]
		}
		return date
	}
	// if value, ok := iso8601time.(string); ok && len(value) >= 10 {
	// 	return value[:10]
	// }
	// Default date
	return "1970-01-01"
}

// String To Date For DB Save.
func StrToDate(datestring any) time.Time {
	switch v := datestring.(type) {
	case string:
		parsedDate, err := time.Parse("2006-01-02", v)
		if err != nil {
			return time.Time{}
		}
		return parsedDate
	case *string:
		if v == nil {
			return time.Time{}
		}
		parsedDate, err := time.Parse("2006-01-02", *v)
		if err != nil {
			return time.Time{}
		}
		return parsedDate
	default:
		return time.Time{}
	}
}
