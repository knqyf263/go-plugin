package timestamppb

import "time"

// Now constructs a new Timestamp from the current time.
func Now() *Timestamp {
	return New(time.Now())
}

// New constructs a new Timestamp from the provided time.Time.
func New(t time.Time) *Timestamp {
	return &Timestamp{Seconds: int64(t.Unix()), Nanos: int32(t.Nanosecond())}
}
