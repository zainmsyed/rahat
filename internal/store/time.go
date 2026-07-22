package store

import (
	"database/sql"
	"time"
)

const (
	TimestampLayout = time.RFC3339
	DateLayout      = "2006-01-02"
)

func FormatTime(t time.Time) string {
	return t.UTC().Format(TimestampLayout)
}

func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

func ParseTime(value string) (time.Time, error) {
	return time.Parse(TimestampLayout, value)
}

func ParseNullableTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}

	parsed, err := ParseTime(value.String)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
