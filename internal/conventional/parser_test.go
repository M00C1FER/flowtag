package conventional

import "testing"

func TestParseFeat(t *testing.T) {
	c := Parse("feat: add user login")
	if !c.IsConventional() {
		t.Fatal("expected conventional")
	}
	if c.Type != "feat" || c.Description != "add user login" {
		t.Errorf("got %+v", c)
	}
	if c.Breaking {
		t.Error("not breaking")
	}
}

func TestParseScopedBreaking(t *testing.T) {
	c := Parse("feat(api)!: drop /v1 endpoints")
	if c.Type != "feat" || c.Scope != "api" || !c.Breaking {
		t.Errorf("got %+v", c)
	}
	if c.Description != "drop /v1 endpoints" {
		t.Errorf("desc=%q", c.Description)
	}
}

func TestParseFooterBreaking(t *testing.T) {
	msg := `feat: rotate tokens

Token format changed.

BREAKING CHANGE: Old tokens stop working at v2 ship time.
Refs: #42`
	c := Parse(msg)
	if !c.Breaking {
		t.Error("BREAKING CHANGE footer should set Breaking=true")
	}
	if len(c.Footers) != 2 {
		t.Errorf("want 2 footers, got %d (%+v)", len(c.Footers), c.Footers)
	}
}

func TestParseNonConventional(t *testing.T) {
	// Per the spec, any noun before `:` qualifies as a type. "WIP" parses,
	// but the bump rules engine treats unknown types as no-bump by default.
	for _, m := range []string{
		"random message no colon",
		"   ",
		"",
	} {
		c := Parse(m)
		if c.IsConventional() {
			t.Errorf("%q should not be conventional, got %+v", m, c)
		}
	}
}

func TestParseEmpty(t *testing.T) {
	c := Parse("")
	if c.IsConventional() {
		t.Error("empty should not be conventional")
	}
}
