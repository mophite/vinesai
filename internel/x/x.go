package x

import (
	"crypto/md5"
	"encoding/hex"
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

func RemoteIp() (string, error) {
	//resp, err := http.Get("https://api.ipify.org?format=text")
	//if err != nil {
	//	return "", err
	//}
	//defer resp.Body.Close()
	//
	//// 读取响应内容
	//body, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	return "", err
	//}
	//
	//return string(body), nil

	return "", nil
}

func Md5(data string) string {
	// Create a new MD5 hash
	hash := md5.New()

	// Write data to it
	hash.Write([]byte(data))

	// Get the resulting hash as a byte slice
	hashBytes := hash.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	return hex.EncodeToString(hashBytes)
}
