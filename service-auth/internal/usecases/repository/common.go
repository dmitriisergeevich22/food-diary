package repository

import (
	"database/sql"
	"time"
)

type IDWrapper struct {
	ID sql.NullInt64 `db:"id"`
}

func BoolToNullBoolean(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}

	return sql.NullBool{Valid: true, Bool: *b}
}

func NullBooleanToBool(b sql.NullBool) *bool {
	if !b.Valid {
		return nil
	}

	return &b.Bool
}

func StringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{Valid: true, String: s}
}

func NullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}

	return ""
}

func IntToNullInt(n int) sql.NullInt64 {
	if n == 0 {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Valid: true, Int64: int64(n)}
}

func NullIntToInt(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}

	return 0
}

func TimeToNullInt(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Valid: true, Time: t}
}

func NullTimeToTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}

	return time.Time{}
}
