package bootstrap

import "testing"

func TestIsJSONContentTypeAcceptsCharset(t *testing.T) {
	if !isJSONContentType("application/json; charset=utf-8") {
		t.Fatal("expected JSON with charset to be accepted")
	}
	if isJSONContentType("text/plain") {
		t.Fatal("text/plain must be rejected")
	}
}
