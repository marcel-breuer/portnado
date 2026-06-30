package domain

import "testing"

func TestDeterministicID(t *testing.T) {
	first := DeterministicID("route", "service_app", "app.webguard.localhost")
	second := DeterministicID("route", "service_app", "app.webguard.localhost")
	other := DeterministicID("route", "service_api", "api.webguard.localhost")
	if first != second {
		t.Fatalf("ids differ: %q != %q", first, second)
	}
	if first == other {
		t.Fatalf("different inputs produced same id: %q", first)
	}
	if len(first) != len("route_")+16 {
		t.Fatalf("id length = %d", len(first))
	}
}
