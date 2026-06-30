package native

import (
	"context"
	"testing"
)

type fakeRunner map[string][]byte

func (f fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	return f[key], nil
}

func TestParseListeners(t *testing.T) {
	listeners, err := parseListeners([]byte("p123\ncnode\nn127.0.0.1:5173\np456\ncpython\nn*:8000\n"))
	if err != nil {
		t.Fatalf("parse listeners: %v", err)
	}
	if len(listeners) != 2 {
		t.Fatalf("listeners = %d", len(listeners))
	}
	if listeners[0].PID != 123 || listeners[0].Command != "node" || listeners[0].Port != 5173 {
		t.Fatalf("listener = %+v", listeners[0])
	}
}

func TestDiscoverSkipsNonLoopback(t *testing.T) {
	detector := Detector{Runner: fakeRunner{
		"lsof -nP -iTCP -sTCP:LISTEN -FpPnac": []byte("p123\ncnode\nn127.0.0.1:5173\np456\ncpython\nn0.0.0.0:8000\n"),
		"lsof -a -p 123 -d cwd -Fn":           []byte("p123\nn/tmp\n"),
	}}
	observations, warnings := detector.Discover(context.Background())
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	if len(observations) != 1 {
		t.Fatalf("observations = %d", len(observations))
	}
	if observations[0].BackendPort != 5173 {
		t.Fatalf("observation = %+v", observations[0])
	}
}
