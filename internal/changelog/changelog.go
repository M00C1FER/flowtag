// Package changelog renders a CHANGELOG entry from a list of conventional commits.
package changelog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/M00C1FER/flowtag/internal/conventional"
)

// Section groups commits by display category.
type Section struct {
	Title   string                 // "Features", "Fixes", "Breaking Changes"
	Commits []conventional.Commit
}

// Render produces a Keep-a-Changelog-style markdown block for one release.
func Render(version string, when time.Time, commits []conventional.Commit) string {
	if when.IsZero() {
		when = time.Now().UTC()
	}
	sections := groupBySection(commits)

	var b strings.Builder
	fmt.Fprintf(&b, "## [%s] - %s\n\n", version, when.Format("2006-01-02"))

	if len(sections) == 0 {
		b.WriteString("_No notable changes._\n")
		return b.String()
	}
	for _, sec := range sections {
		fmt.Fprintf(&b, "### %s\n\n", sec.Title)
		for _, c := range sec.Commits {
			line := c.Description
			if c.Scope != "" {
				line = fmt.Sprintf("**%s:** %s", c.Scope, line)
			}
			fmt.Fprintf(&b, "- %s", line)
			if c.SHA != "" {
				fmt.Fprintf(&b, " (`%s`)", c.SHA)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func groupBySection(commits []conventional.Commit) []Section {
	type key struct{ order int; title string }
	titleFor := func(c conventional.Commit) (key, bool) {
		if c.Breaking {
			return key{0, "Breaking Changes"}, true
		}
		switch c.Type {
		case "feat":
			return key{1, "Features"}, true
		case "fix":
			return key{2, "Fixes"}, true
		case "perf":
			return key{3, "Performance"}, true
		case "refactor":
			return key{4, "Refactors"}, true
		case "docs":
			return key{5, "Documentation"}, true
		case "test":
			return key{6, "Tests"}, true
		case "ci", "build", "chore":
			return key{7, "Maintenance"}, true
		case "revert":
			return key{8, "Reverts"}, true
		}
		return key{}, false
	}
	groups := make(map[key][]conventional.Commit)
	for _, c := range commits {
		if !c.IsConventional() {
			continue
		}
		if k, ok := titleFor(c); ok {
			groups[k] = append(groups[k], c)
		}
	}
	keys := make([]key, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].order < keys[j].order })

	out := make([]Section, 0, len(keys))
	for _, k := range keys {
		out = append(out, Section{Title: k.title, Commits: groups[k]})
	}
	return out
}
