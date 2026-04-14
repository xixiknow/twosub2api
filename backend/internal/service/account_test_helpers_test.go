//go:build unit

package service

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func newAccountTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/test", nil)
	return ctx, recorder
}

func newJSONResponse(status int, body string) *http.Response {
	respBody := body
	if respBody == "" {
		respBody = "{}"
	}
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(respBody)),
	}
}

type queuedHTTPUpstream struct {
	responses []*http.Response
	errors    []error
	callCount int
}

func (u *queuedHTTPUpstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	idx := u.callCount
	u.callCount++

	var resp *http.Response
	if idx < len(u.responses) {
		resp = u.responses[idx]
	}

	var err error
	if idx < len(u.errors) {
		err = u.errors[idx]
	}

	if resp == nil && err == nil {
		return nil, errors.New("unexpected upstream call")
	}
	return resp, err
}

func (u *queuedHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, concurrency int, _ bool) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, concurrency)
}
