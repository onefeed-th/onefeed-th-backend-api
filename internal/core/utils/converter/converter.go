package converter

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func StringPointerToPGTypeTextNull(s *string) pgtype.Text {
	return pgtype.Text{
		Valid:  s != nil,
		String: *s,
	}
}

func StringToPGTypeTextNull(s string) pgtype.Text {
	return pgtype.Text{
		Valid:  s != "",
		String: s,
	}
}

func TimePointerToPGTypeTimestamp(s *time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Valid: s != nil,
		Time:  *s,
	}
}
