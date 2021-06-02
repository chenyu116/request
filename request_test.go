package request

import (
	"testing"
)

func TestSimple(t *testing.T) {
	statusCode, err := NewRequest("http://example.com").Do()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("statusCode: %d", statusCode)
}

func TestWithQuery(t *testing.T) {
	r := NewRequest("http://example.com", WithQuery("key1", "value",""))
	statusCode, err := r.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("statusCode: %d", statusCode)
	t.Logf("request path: %v", r.GetRequest().URL)
}
