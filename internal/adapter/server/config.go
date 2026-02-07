package server

import "time"

type HttpConfig interface {
	Host() string
	Port() int
	ReadHeaderTimeout() time.Duration
	ReadTimeout() time.Duration
	IdleTimeout() time.Duration
	HandlerTimeout() time.Duration

	ShutDownTimeout() time.Duration
	ShutDownHardTimeout() time.Duration
	ShutDownReadyDelay() time.Duration
}
