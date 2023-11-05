package lib

import (
	"time"
)

// During convert integer and time types
func During(i int) time.Duration {
	return time.Duration(i)
}
