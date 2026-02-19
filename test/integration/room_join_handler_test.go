package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dyxj/chess/internal/engine"
	"github.com/dyxj/chess/internal/room"
	"github.com/dyxj/chess/pkg/httpx"
	"github.com/dyxj/chess/pkg/randx"
	"github.com/dyxj/chess/test/testx"
	"github.com/stretchr/testify/assert"
)

func TestRoomJoinHandler(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	r := room.NewEmptyRoom()
	err := memCache.Add(r.Code, r, time.Time{})
	assert.NoError(t, err)

	color := randx.FromSlice(engine.Colors)
	payload := room.JoinRequest{
		Name:  color.String() + " player",
		Color: color,
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(payload)
	assert.NoError(t, err)

	request, err := http.NewRequest(
		"POST",
		testSvr.URL+fmt.Sprintf("/room/%s/join", r.Code),
		buffer)
	assert.NoError(t, err)

	resp, err := testSvr.Client().Do(request)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result room.JoinResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.NotEmpty(t, result.Token)
}

func TestRoomJoinHandler_ShouldReturnBadRequest_InvalidRequestBody(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	r := room.NewEmptyRoom()
	err := memCache.Add(r.Code, r, time.Time{})
	assert.NoError(t, err)

	request, err := http.NewRequest(
		"POST",
		testSvr.URL+fmt.Sprintf("/room/%s/join", r.Code),
		// invalid JSON: without {}
		bytes.NewBuffer([]byte(`name: "test player", color: "white"`)))
	assert.NoError(t, err)

	resp, err := testSvr.Client().Do(request)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result httpx.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, httpx.CodeBadRequest, result.Code)
	assert.Equal(t, "invalid request body", result.Message)
	assert.Equal(t,
		"invalid character 'a' in literal null (expecting 'u')",
		result.Details["error"],
	)
}

func TestRoomJoinHandler_ShouldReturnBadRequest_InvalidCode(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	code := "invalid"

	color := randx.FromSlice(engine.Colors)
	payload := room.JoinRequest{
		Name:  color.String() + " player",
		Color: color,
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(payload)
	assert.NoError(t, err)

	request, err := http.NewRequest(
		"POST",
		testSvr.URL+fmt.Sprintf("/room/%s/join", code),
		buffer)
	assert.NoError(t, err)

	resp, err := testSvr.Client().Do(request)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result httpx.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, httpx.CodeBadRequest, result.Code)
	assert.Equal(t, "validation failed", result.Message)
	assert.Equal(t,
		"code length must be 6 characters",
		result.Details["code"],
	)
}

func TestRoomJoinHandler_ShouldReturnBadRequest_PayloadValidation(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	r := room.NewEmptyRoom()
	err := memCache.Add(r.Code, r, time.Time{})
	assert.NoError(t, err)

	tt := []struct {
		name       string
		payload    func(room.JoinRequest) room.JoinRequest
		expMsg     string
		expProp    string
		expPropMsg string
	}{
		{
			name: "empty name",
			payload: func(r room.JoinRequest) room.JoinRequest {
				r.Name = ""
				return r
			},
			expMsg:     "validation failed",
			expProp:    "name",
			expPropMsg: "name is required",
		},
		{
			name: "invalid color",
			payload: func(r room.JoinRequest) room.JoinRequest {
				r.Color = engine.Color(0)
				return r
			},
			expMsg:     "invalid request body",
			expProp:    "error",
			expPropMsg: "unknown color: unknown valid color(white,black)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			color := randx.FromSlice(engine.Colors)
			payload := room.JoinRequest{
				Name:  color.String() + " player",
				Color: color,
			}

			buffer := new(bytes.Buffer)
			err = json.NewEncoder(buffer).Encode(tc.payload(payload))
			assert.NoError(t, err)

			request, err := http.NewRequest(
				"POST",
				testSvr.URL+fmt.Sprintf("/room/%s/join", r.Code),
				buffer)
			assert.NoError(t, err)

			resp, err := testSvr.Client().Do(request)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			var result httpx.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)

			assert.Equal(t, httpx.CodeBadRequest, result.Code)
			assert.Equal(t, tc.expMsg, result.Message)
			assert.Equal(t, tc.expPropMsg, result.Details[tc.expProp])
		})
	}
}

func TestRoomJoinHandler_ShouldReturnNotFound(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	color := randx.FromSlice(engine.Colors)
	payload := room.JoinRequest{
		Name:  color.String() + " player",
		Color: color,
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(payload)
	assert.NoError(t, err)

	request, err := http.NewRequest(
		"POST",
		testSvr.URL+fmt.Sprintf("/room/%s/join", "A12B3C"),
		buffer)
	assert.NoError(t, err)

	resp, err := testSvr.Client().Do(request)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result httpx.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, httpx.CodeEntityNotFound, result.Code)
	assert.Equal(t, "entity not found", result.Message)
}

func TestRoomJoinHandler_ShouldReturnBadRequest_MaxTicketsIssued(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	var err error
	r := room.NewEmptyRoom()
	err = memCache.Add(r.Code, r, time.Time{})
	assert.NoError(t, err)

	rCoordinator := testx.GlobalEnv().RoomCoordinator()
	_, err = rCoordinator.IssueTicketToken(r.Code, "white player", engine.White)
	_, err = rCoordinator.IssueTicketToken(r.Code, "black player", engine.Black)

	color := randx.FromSlice(engine.Colors)
	payload := room.JoinRequest{
		Name:  color.String() + " player",
		Color: color,
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(payload)
	assert.NoError(t, err)

	request, err := http.NewRequest(
		"POST",
		testSvr.URL+fmt.Sprintf("/room/%s/join", r.Code),
		buffer)
	assert.NoError(t, err)

	resp, err := testSvr.Client().Do(request)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result httpx.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, httpx.CodeBadRequest, result.Code)
	assert.Equal(t, "room is full", result.Message)
}

func TestRoomJoinHandler_ShouldReturnBadRequest_RoomStatusNotWaiting(t *testing.T) {
	testSvr := testx.GlobalEnv().HTTTPTestServer()

	memCache := testx.GlobalEnv().MemCache()
	t.Cleanup(func() {
		memCache.Clear()
	})

	var err error
	r := room.NewEmptyRoom()
	err = memCache.Add(r.Code, r, time.Time{})
	assert.NoError(t, err)

	tt := []struct {
		status room.Status
	}{
		{status: room.StatusInProgress},
		{status: room.StatusCompleted},
	}

	for _, tc := range tt {
		t.Run(tc.status.String(), func(t *testing.T) {
			r.SetStatus(tc.status)

			color := randx.FromSlice(engine.Colors)
			payload := room.JoinRequest{
				Name:  color.String() + " player",
				Color: color,
			}

			buffer := new(bytes.Buffer)
			err = json.NewEncoder(buffer).Encode(payload)
			assert.NoError(t, err)

			request, err := http.NewRequest(
				"POST",
				testSvr.URL+fmt.Sprintf("/room/%s/join", r.Code),
				buffer)
			assert.NoError(t, err)

			resp, err := testSvr.Client().Do(request)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			var result httpx.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)

			assert.Equal(t, httpx.CodeBadRequest, result.Code)
			assert.Equal(t, "room is full", result.Message)
		})
	}
}
