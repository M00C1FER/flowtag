package conventional

import (
	"strings"
	"testing"
)

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

// --- edge-case tests added during audit cycle ---

// TestParseBodyLinesNotMisclassifiedAsFooters verifies that body text
// containing words followed by ": " is NOT treated as a footer token when
// those lines appear as part of a multi-word sentence (not a bare token).
func TestParseBodyLinesNotMisclassifiedAsFooters(t *testing.T) {
	msg := "fix: correct parser\n\nThis fix addresses how Doc: annotations are handled.\nPreviously Refs: broken."
	c := Parse(msg)
	if len(c.Footers) != 0 {
		t.Errorf("body lines should not become footers; got %+v", c.Footers)
	}
	if !strings.Contains(c.Body, "Doc: annotations") {
		t.Errorf("body text lost (Doc line); body=%q", c.Body)
	}
	if !strings.Contains(c.Body, "Refs: broken") {
		t.Errorf("body text lost (Refs line); body=%q", c.Body)
	}
}

// TestParseMultiParagraphBodyNoFooter verifies that a commit with multiple
// body paragraphs and no footer block is parsed with all text as body.
func TestParseMultiParagraphBodyNoFooter(t *testing.T) {
	msg := "fix: something\n\nFirst paragraph.\n\nSecond paragraph has non-footer text here."
	c := Parse(msg)
	if len(c.Footers) != 0 {
		t.Errorf("no footers expected; got %+v", c.Footers)
	}
	if !strings.Contains(c.Body, "Second paragraph") {
		t.Errorf("body missing second paragraph; body=%q", c.Body)
	}
}

// TestParseHashTokenFooter tests the git-trailer "token #value" footer form
// (e.g. "Closes #101") defined in the CC spec.
func TestParseHashTokenFooter(t *testing.T) {
	msg := "fix: handle edge case\n\nSmall fix.\n\nCloses #101\nFixes #202"
	c := Parse(msg)
	if len(c.Footers) != 2 {
		t.Errorf("want 2 footers, got %d: %+v", len(c.Footers), c.Footers)
	}
	found := false
	for _, f := range c.Footers {
		if f.Key == "Closes" && f.Value == "#101" {
			found = true
		}
	}
	if !found {
		t.Errorf("Closes #101 footer not found: %+v", c.Footers)
	}
}

// TestParseBodyFooterBoundary verifies that body text containing footer-shaped
// tokens is correctly preserved when a real footer block follows a later blank line.
func TestParseBodyFooterBoundary(t *testing.T) {
	msg := "feat: api\n\nThis is body.\nSome context: still body.\n\nBREAKING CHANGE: real footer\nRefs: #9"
	c := Parse(msg)
	if !c.Breaking {
		t.Error("BREAKING CHANGE footer should set Breaking=true")
	}
	if len(c.Footers) != 2 {
		t.Errorf("want 2 footers, got %d: %+v", len(c.Footers), c.Footers)
	}
	if !strings.Contains(c.Body, "Some context: still body") {
		t.Errorf("body text lost; body=%q", c.Body)
	}
}

// TestParseFooterRequiresBlankLine verifies that a BREAKING CHANGE marker
// immediately following the subject (without a preceding blank line) is treated
// as body text per the CC spec — footers MUST be preceded by a blank line.
func TestParseFooterRequiresBlankLine(t *testing.T) {
	msg := "feat: new api\nBREAKING CHANGE: no blank line before this"
	c := Parse(msg)
	if c.Breaking {
		t.Error("BREAKING CHANGE without preceding blank line should not be a footer")
	}
	if len(c.Footers) != 0 {
		t.Errorf("expected no footers; got %+v", c.Footers)
	}
}

// TestParseBreakingChangeHyphen verifies that the BREAKING-CHANGE token
// (hyphen form) is also recognized as a breaking-change footer.
func TestParseBreakingChangeHyphen(t *testing.T) {
	msg := "feat: new api\n\nBREAKING-CHANGE: old API removed"
	c := Parse(msg)
	if !c.Breaking {
		t.Error("BREAKING-CHANGE footer should set Breaking=true")
	}
}
