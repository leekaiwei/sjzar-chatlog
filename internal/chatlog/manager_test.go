package chatlog

import "testing"

func TestValidateHTTPAddrRejectsWildcard(t *testing.T) {
	if err := validateHTTPAddr("0.0.0.0:5030"); err == nil {
		t.Fatal("expected wildcard bind address to be rejected")
	}
}

func TestValidateHTTPAddrAllowsLoopback(t *testing.T) {
	if err := validateHTTPAddr("127.0.0.1:5030"); err != nil {
		t.Fatalf("expected loopback bind address to be allowed: %v", err)
	}
}
