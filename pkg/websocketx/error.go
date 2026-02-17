package websocketx

import (
	"errors"
	"fmt"
	"net"

	"github.com/gobwas/ws/wsutil"
)

func IsNetworkClosedError(err error) bool {
	return errors.Is(err, net.ErrClosed)
}

func IsWebSocketClosedError(err error) (wsutil.ClosedError, bool) {
	var closeErr wsutil.ClosedError
	if errors.As(err, &closeErr) {
		return closeErr, true
	}
	return wsutil.ClosedError{}, false
}

type InvalidPayloadError struct {
	Msg string
	Err error
}

func (e *InvalidPayloadError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *InvalidPayloadError) Unwrap() error {
	return e.Err
}
