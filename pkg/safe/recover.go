package safe

import (
	"log"
	"runtime/debug"

	"go.uber.org/zap"
)

func RecoverWithLog(logger *zap.Logger, msg string) func() {
	return func() {
		if err := recover(); err != nil {
			logger.Error(msg,
				zap.Any("panic", err),
			)
			debug.PrintStack()
		}
	}
}

func RecoverFn(fn func(err interface{}, debugStack []byte)) func() {
	return func() {
		if err := recover(); err != nil {
			fn(err, debug.Stack())
		}
	}
}

func Recover() func() {
	return func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from goroutine panic: %v\n%s", err, debug.Stack())
		}
	}
}
