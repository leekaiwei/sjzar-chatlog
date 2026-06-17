package http

import "testing"

func TestHostHeaderAllowedRejectsReboundHost(t *testing.T) {
	allowed, err := hostHeaderAllowed("attacker.test:5030", "127.0.0.1:5030")
	if err != nil {
		t.Fatalf("hostHeaderAllowed returned error: %v", err)
	}
	if allowed {
		t.Fatal("expected non-localhost Host header to be rejected")
	}
}

func TestHostHeaderAllowedAllowsLoopbackAndLocalhost(t *testing.T) {
	for _, host := range []string{"127.0.0.1:5030", "localhost:5030", "[::1]:5030"} {
		allowed, err := hostHeaderAllowed(host, "127.0.0.1:5030")
		if err != nil {
			t.Fatalf("hostHeaderAllowed(%q) returned error: %v", host, err)
		}
		if !allowed {
			t.Fatalf("expected %q to be allowed", host)
		}
	}
}

func TestHostHeaderAllowedRejectsWrongPort(t *testing.T) {
	allowed, err := hostHeaderAllowed("127.0.0.1:80", "127.0.0.1:5030")
	if err != nil {
		t.Fatalf("hostHeaderAllowed returned error: %v", err)
	}
	if allowed {
		t.Fatal("expected wrong Host port to be rejected")
	}
}

func TestHostHeaderAllowedPortlessOnPort80(t *testing.T) {
	// HTTP clients omit the port in Host header when binding to port 80.
	allowed, err := hostHeaderAllowed("127.0.0.1", "127.0.0.1:80")
	if err != nil {
		t.Fatalf("hostHeaderAllowed returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected portless loopback Host header to be allowed on port 80")
	}
}

func TestHostHeaderAllowedPortlessRejectsNonLoopback(t *testing.T) {
	// Portless non-loopback host should be rejected even with port-80 server.
	allowed, err := hostHeaderAllowed("attacker.test", "127.0.0.1:80")
	if err != nil {
		t.Fatalf("hostHeaderAllowed returned error: %v", err)
	}
	if allowed {
		t.Fatal("expected portless non-loopback Host header to be rejected")
	}
}

func TestHostHeaderAllowedPortlessMismatchPort(t *testing.T) {
	// Portless Host (implicit 80) should be rejected when server is on a different port.
	allowed, err := hostHeaderAllowed("127.0.0.1", "127.0.0.1:5030")
	if err != nil {
		t.Fatalf("hostHeaderAllowed returned error: %v", err)
	}
	if allowed {
		t.Fatal("expected portless Host header to be rejected when server is not on port 80")
	}
}
