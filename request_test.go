package request

import (
	"testing"
)

func TestSimple(t *testing.T) {
	statusCode, err := NewRequest("http://baidu.com").Do()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("statusCode: %d", statusCode)
}

func TestWithQuery(t *testing.T) {
	r := NewRequest("http://baidu.com", WithQuery("key1", "value", ""))
	statusCode, err := r.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("statusCode: %d", statusCode)
	t.Logf("request path: %v", r.GetRequest().URL)
}
func TestDo(t *testing.T) {
	r, err := Do("http://baidu.com", WithQuery("key1", "value", ""))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("statusCode: %d", r.StatusCode())
	t.Logf("request path: %v", r.GetRequest().URL)
}
