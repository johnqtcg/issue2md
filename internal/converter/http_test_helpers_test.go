package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func mustJSONResponse(status int, payload any) *http.Response {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("marshal response payload: %v", err))
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
	}
}
