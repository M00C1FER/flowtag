package bump

import (
	"testing"

	"github.com/M00C1FER/flowtag/internal/conventional"
)

func parse(t *testing.T, msgs ...string) []conventional.Commit {
	t.Helper()
	out := make([]conventional.Commit, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, conventional.Parse(m))
	}
	return out
}

func TestComputeFeatIsMinor(t *testing.T) {
	commits := parse(t, "feat: a", "fix: b")
	if got := Compute(commits, DefaultRules()); got != Minor {
		t.Errorf("got %s want Minor", got)
	}
}

func TestComputeBreakingIsMajor(t *testing.T) {
	commits := parse(t, "fix: a", "feat!: b")
	if got := Compute(commits, DefaultRules()); got != Major {
		t.Errorf("got %s want Major", got)
	}
}

func TestComputeFooterBreakingIsMajor(t *testing.T) {
	commits := parse(t, "feat: regular\n\nBREAKING CHANGE: API removed")
	if got := Compute(commits, DefaultRules()); got != Major {
		t.Errorf("got %s want Major", got)
	}
}

func TestComputeOnlyFixIsPatch(t *testing.T) {
	commits := parse(t, "fix: a", "fix: b", "chore: c")
	if got := Compute(commits, DefaultRules()); got != Patch {
		t.Errorf("got %s want Patch", got)
	}
}

func TestComputeOnlyChoreIsNone(t *testing.T) {
	commits := parse(t, "chore: a", "docs: b", "style: c")
	if got := Compute(commits, DefaultRules()); got != None {
		t.Errorf("got %s want None", got)
	}
}

func TestApplyVersions(t *testing.T) {
	cases := []struct {
		cur  string
		kind Kind
		want string
	}{
		{"v0.1.0", Patch, "v0.1.1"},
		{"v0.1.0", Minor, "v0.2.0"},
		{"v0.1.5", Major, "v1.0.0"},
		{"1.2.3", Patch, "1.2.4"},
		{"1.2.3-rc1", Minor, "1.3.0"},
		{"v0.0.0", Patch, "v0.0.1"},
		{"", Minor, "0.1.0"},
	}
	for _, c := range cases {
		got, err := Apply(c.cur, c.kind)
		if err != nil {
			t.Errorf("Apply(%q, %s) err: %v", c.cur, c.kind, err)
			continue
		}
		if got != c.want {
			t.Errorf("Apply(%q, %s) = %q want %q", c.cur, c.kind, got, c.want)
		}
	}
}

func TestApplyNoneIsIdentity(t *testing.T) {
	got, err := Apply("v0.1.0", None)
	if err != nil || got != "v0.1.0" {
		t.Errorf("got %q,%v want v0.1.0,nil", got, err)
	}
}
