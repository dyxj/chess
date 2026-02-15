package websocketx

import (
	"errors"
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
