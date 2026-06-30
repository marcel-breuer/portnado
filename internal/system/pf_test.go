package system

import (
	"strings"
	"testing"
)

func TestRenderPFAnchor(t *testing.T) {
	anchor, err := RenderPFAnchor(4780)
	if err != nil {
		t.Fatalf("render PF anchor: %v", err)
	}
	if !strings.Contains(anchor, "127.0.0.1 port 80") {
		t.Fatalf("anchor missing source port: %s", anchor)
	}
	if !strings.Contains(anchor, "127.0.0.1 port 4780") {
		t.Fatalf("anchor missing proxy port: %s", anchor)
	}
	if strings.Contains(anchor, "{{PORT}}") {
		t.Fatalf("anchor still contains template marker: %s", anchor)
	}
}

func TestRenderPFAnchorRejectsInvalidPort(t *testing.T) {
	if _, err := RenderPFAnchor(0); err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestManagedPFIncludeLine(t *testing.T) {
	if got := ManagedPFIncludeLine(); got != `rdr-anchor "dev.portnado"` {
		t.Fatalf("include line = %q", got)
	}
}
