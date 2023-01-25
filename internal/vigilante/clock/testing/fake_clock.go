package testing

import "time"

type FakeClock struct {
	CurrentTime time.Time
}

func (f *FakeClock) Now() time.Time {
	return f.CurrentTime
}

func (f *FakeClock) PassTime(duration time.Duration) {
	f.CurrentTime = f.CurrentTime.Add(duration)
}
