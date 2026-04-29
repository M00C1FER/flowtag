package changelog

import (
	"strings"
	"testing"
	"time"

	"github.com/M00C1FER/flowtag/internal/conventional"
)

func TestRenderHasSections(t *testing.T) {
	commits := []conventional.Commit{
		conventional.Parse("feat: add login"),
		conventional.Parse("fix: handle empty header"),
		conventional.Parse("feat(api)!: drop /v1"),
	}
	out := Render("v0.2.0", time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC), commits)

	for _, want := range []string{
		"## [v0.2.0] - 2026-04-29",
		"### Breaking Changes",
		"### Features",
		"### Fixes",
		"add login",
		"handle empty header",
		"drop /v1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in:\n%s", want, out)
		}
	}
}

func TestRenderEmpty(t *testing.T) {
	out := Render("v0.1.0", time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC), nil)
	if !strings.Contains(out, "No notable changes") {
		t.Errorf("empty changelog should mark it; got:\n%s", out)
	}
}

func TestRenderScopePrefix(t *testing.T) {
	commits := []conventional.Commit{conventional.Parse("feat(auth): add token rotation")}
	out := Render("v0.2.0", time.Now(), commits)
	if !strings.Contains(out, "**auth:** add token rotation") {
		t.Errorf("expected scope prefix; got:\n%s", out)
	}
}
