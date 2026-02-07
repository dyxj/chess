package httpx

import (
	"net/http"
	"time"
)

const timeoutResponseBody = `{message: "request timeout"}`

func TimeoutHandler(h http.Handler, timeout time.Duration) http.Handler {
	return http.TimeoutHandler(h, timeout, timeoutResponseBody)
}
