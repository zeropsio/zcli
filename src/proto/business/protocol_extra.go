package business

import (
	"time"
)

func ToProtoTimestamp(t time.Time) *Timestamp {
	if t.IsZero() {
		return &Timestamp{}
	}
	return &Timestamp{IsSet: true, Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}

func FromProtoTimestamp(t *Timestamp) time.Time {
	if !t.GetIsSet() {
		return time.Time{}
	}
	return time.Unix(t.GetSeconds(), int64(t.GetNanos()))
}
