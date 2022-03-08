package request

import (
	"testing"
)

func TestSimple(t *testing.T) {
	r, err := GET("http://baidu.com")
	if err != nil {
		t.Error(err)
	}

	t.Logf("statusCode: %d", r.StatusCode())
}

func TestWithQuery(t *testing.T) {
	r, err := GET("http://baidu.com", WithQuery("key1", "value", ""))
	if err != nil {
		t.Error(err)
	}
	t.Logf("statusCode: %d", r.StatusCode())
	t.Logf("request path: %v", r.Request().URL)
}
