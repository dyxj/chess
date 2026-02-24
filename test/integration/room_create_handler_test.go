package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/dyxj/chess/pkg/room"
	"github.com/dyxj/chess/test/testx"
	"github.com/stretchr/testify/assert"
)

func TestRoomCreateHandler(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	request, err := http.NewRequest(
		"POST",
		testSvr.URL+"/room",
		nil)
	assert.NoError(t, err)

	resp, err := testSvr.Client().Do(request)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result room.CreateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Len(t, result.Code, 6)
	assert.Equal(t, "waiting", result.Status)
	assert.WithinDuration(t, time.Now(), result.CreatedTime, 5*time.Second)
}
