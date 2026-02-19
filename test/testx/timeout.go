package testx

import (
	"time"
)

func SetTimeout(timeout time.Duration) {
	<-time.After(timeout)
	panic("test timed out")
}
