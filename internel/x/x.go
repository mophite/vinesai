package x

import (
	"time"

	"github.com/RussellLuo/timingwheel"
)

const (
	StatusOK                  = "success"
	StatusInternalServerError = "internal server error"
	StatusBadRequest          = "bad request"
)

var tw = timingwheel.NewTimingWheel(time.Second, 20)

func init() {
	tw.Start()
}

type EveryScheduler struct {
	Interval time.Duration
}

func (s *EveryScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}

func TimingwheelAfter(t time.Duration, f func()) {
	tw.AfterFunc(t, f)
}

func TimingwheelTicker(t time.Duration, f func()) *timingwheel.Timer {
	return tw.ScheduleFunc(&EveryScheduler{Interval: t}, f)
}
