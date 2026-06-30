package doctor

import (
	"context"
	"testing"
	"time"
)

func TestRunIncludesExpectedChecks(t *testing.T) {
	report := Run(context.Background(), Options{
		ProxyAddress: "127.0.0.1:1",
		LookupHost:   "localhost",
		Now:          func() time.Time { return time.Unix(10, 0) },
	})
	if got, want := report.GeneratedAt, time.Unix(10, 0).UTC(); !got.Equal(want) {
		t.Fatalf("GeneratedAt = %s, want %s", got, want)
	}
	ids := map[string]bool{}
	for _, check := range report.Checks {
		ids[check.ID] = true
	}
	for _, id := range []string{"platform", "localhost-resolution", "control-socket", "sqlite", "http-proxy", "docker", "launch-agent"} {
		if !ids[id] {
			t.Fatalf("missing check %q in %#v", id, report.Checks)
		}
	}
}

func TestReportHasFailures(t *testing.T) {
	report := Report{Checks: []Check{{ID: "ok", Status: StatusPass}, {ID: "bad", Status: StatusFail}}}
	if !report.HasFailures() {
		t.Fatal("expected report to detect failures")
	}
}

func TestReportHasFailuresFalse(t *testing.T) {
	report := Report{Checks: []Check{{ID: "ok", Status: StatusPass}, {ID: "warn", Status: StatusWarn}}}
	if report.HasFailures() {
		t.Fatal("did not expect failures")
	}
}
