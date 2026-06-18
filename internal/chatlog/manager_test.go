package chatlog

import "testing"

func TestValidateHTTPAddrRejectsWildcard(t *testing.T) {
	for _, addr := range []string{"0.0.0.0:5030", "[::]:5030", ":5030"} {
		if err := validateHTTPAddr(addr); err == nil {
			t.Fatalf("expected %q to be rejected as a non-loopback bind address", addr)
		}
	}
}

func TestValidateHTTPAddrAllowsLoopback(t *testing.T) {
	if err := validateHTTPAddr("127.0.0.1:5030"); err != nil {
		t.Fatalf("expected loopback bind address to be allowed: %v", err)
	}
}
