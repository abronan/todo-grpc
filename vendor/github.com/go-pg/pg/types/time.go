package types

import (
	"time"

	"github.com/go-pg/pg/internal"
	"github.com/gogo/protobuf/types"
)

const (
	dateFormat         = "2006-01-02"
	timeFormat         = "15:04:05.999999999"
	timestampFormat    = "2006-01-02 15:04:05.999999999"
	timestamptzFormat  = "2006-01-02 15:04:05.999999999-07:00:00"
	timestamptzFormat2 = "2006-01-02 15:04:05.999999999-07:00"
	timestamptzFormat3 = "2006-01-02 15:04:05.999999999-07"
)

func ParseTime(b []byte) (time.Time, error) {
	s := internal.BytesToString(b)
	switch l := len(b); {
	case l <= len(dateFormat):
		return time.ParseInLocation(dateFormat, s, time.UTC)
	case l <= len(timeFormat):
		return time.ParseInLocation(timeFormat, s, time.UTC)
	default:
		if c := b[len(b)-9]; c == '+' || c == '-' {
			return time.Parse(timestamptzFormat, s)
		}
		if c := b[len(b)-6]; c == '+' || c == '-' {
			return time.Parse(timestamptzFormat2, s)
		}
		if c := b[len(b)-3]; c == '+' || c == '-' {
			return time.Parse(timestamptzFormat3, s)
		}
		return time.ParseInLocation(timestampFormat, s, time.UTC)
	}
}

func AppendTime(b []byte, tm time.Time, quote int) []byte {
	if quote == 1 {
		b = append(b, '\'')
	}
	b = tm.UTC().AppendFormat(b, timestamptzFormat)
	if quote == 1 {
		b = append(b, '\'')
	}
	return b
}

func AppendGrpcTime(b []byte, ts types.Timestamp, quote int) []byte {
	if quote == 1 {
		b = append(b, '\'')
	}
	tm, err := types.TimestampFromProto(&ts)
	if err != nil {
		return nil
	}
	b = tm.UTC().AppendFormat(b, timestamptzFormat)
	if quote == 1 {
		b = append(b, '\'')
	}
	return b
}
