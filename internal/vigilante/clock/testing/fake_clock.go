package testing

import "time"

type FakeClock struct {
	CurrentTime time.Time
}

func (f *FakeClock) Now() time.Time {
	return f.CurrentTime
}
