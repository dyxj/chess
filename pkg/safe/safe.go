package safe

import (
	"log"
	"runtime/debug"

	"go.uber.org/zap"
)

func Go(fn func()) {
	GoWithRecover(fn, nil)
}

func GoWithLog(fn func(), logger *zap.Logger, msg string) {
	GoWithRecover(fn, func(err interface{}, debugStack []byte) {
		logger.Error(msg,
			zap.Any("panic", err),
			zap.ByteString("stack", debugStack),
		)
	})
}

func GoWithRecover(
	fn func(),
	recoverFn func(err interface{}, debugStack []byte),
) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if recoverFn != nil {
					recoverFn(err, debug.Stack())
				} else {
					log.Printf("Recovered from goroutine panic: %v\n%s", err, debug.Stack())
				}
			}
		}()
		fn()
	}()
}
