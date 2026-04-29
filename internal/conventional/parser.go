// Package conventional parses Conventional Commits messages.
//
// Spec: https://www.conventionalcommits.org/en/v1.0.0/
//
// Subject form:   <type>(<scope>)?(!)?: <description>
// Body footer:    BREAKING CHANGE: <text>      (also: Refs:, Closes:, etc.)
package conventional

import (
	"regexp"
	"strings"
)

// Commit is one parsed conventional commit.
type Commit struct {
	SHA         string   // short hash, optional
	Type        string   // "feat", "fix", "chore", etc.
	Scope       string   // optional, e.g., "auth"
	Description string   // post-colon subject
	Body        string   // raw body (excluding subject + footers)
	Breaking    bool     // either `!` after type/scope OR a BREAKING CHANGE: footer
	Footers     []Footer // parsed footer lines
	Raw         string   // original full message
}

// Footer is a key-value line below the body, e.g., "Refs: #42".
type Footer struct {
	Key, Value string
}

var (
	// `<type>(<scope>)?(!)?:`  with optional scope and breaking marker
	headerRe = regexp.MustCompile(`^([a-zA-Z]+)(\(([^)]+)\))?(!)?:\s*(.+)$`)
	// Footer: "Key: value" or "Key #value"
	footerRe = regexp.MustCompile(`^([A-Z][A-Za-z-]+|BREAKING CHANGE):\s*(.+)$`)
)

// Parse parses one full commit message into a Commit. Non-conventional messages
// return a Commit with Type="" — caller decides how to treat them.
func Parse(message string) Commit {
	c := Commit{Raw: message}
	lines := strings.Split(strings.ReplaceAll(message, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return c
	}
	subject := strings.TrimSpace(lines[0])
	if m := headerRe.FindStringSubmatch(subject); m != nil {
		c.Type = strings.ToLower(m[1])
		c.Scope = m[3]
		c.Breaking = m[4] == "!"
		c.Description = strings.TrimSpace(m[5])
	}
	// Body + footers
	if len(lines) > 1 {
		var bodyLines []string
		for _, line := range lines[1:] {
			line = strings.TrimRight(line, " \t")
			if m := footerRe.FindStringSubmatch(line); m != nil {
				key, val := m[1], m[2]
				c.Footers = append(c.Footers, Footer{Key: key, Value: val})
				if key == "BREAKING CHANGE" || key == "BREAKING-CHANGE" {
					c.Breaking = true
				}
				continue
			}
			bodyLines = append(bodyLines, line)
		}
		c.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}
	return c
}

// IsConventional returns true if Parse extracted a recognizable type.
func (c Commit) IsConventional() bool { return c.Type != "" }
