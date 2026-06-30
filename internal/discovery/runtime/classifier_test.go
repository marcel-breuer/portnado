package runtime

import "testing"

func TestClassifyNodeVite(t *testing.T) {
	got := Classify("node", []string{"vite", "--host", "127.0.0.1"}, []string{"package.json"})
	if got.Runtime != "node" || got.Service != "app" || got.Confidence != "high" {
		t.Fatalf("classification = %+v", got)
	}
}

func TestClassifyUnknown(t *testing.T) {
	got := Classify("custom-server", nil, nil)
	if got.Runtime != "unknown" || got.Confidence != "low" {
		t.Fatalf("classification = %+v", got)
	}
}
