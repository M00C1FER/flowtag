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
	// Footer colon form:  "Token: value"  or  "BREAKING CHANGE: value"
	// Tokens must start with an uppercase letter per the git-trailer convention.
	footerColonRe = regexp.MustCompile(`^([A-Z][A-Za-z-]*|BREAKING CHANGE):\s*(.+)$`)
	// Footer hash form (git-trailer style): "Token #value"  e.g. "Fixes #42"
	footerHashRe = regexp.MustCompile(`^([A-Z][A-Za-z-]*) #(.+)$`)
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
	//
	// Per CC spec §12, footers MUST begin after a blank line that terminates
	// the body.  We locate the footer block by finding the last blank-line
	// boundary after which every non-empty line is a valid footer token.
	if len(lines) > 1 {
		rest := lines[1:]
		footerStart := findFooterStart(rest)
		bodyLines := rest
		if footerStart >= 0 {
			bodyLines = rest[:footerStart]
			for _, line := range rest[footerStart:] {
				line = strings.TrimRight(line, " \t")
				if line == "" {
					continue
				}
				if m := footerColonRe.FindStringSubmatch(line); m != nil {
					key, val := m[1], m[2]
					c.Footers = append(c.Footers, Footer{Key: key, Value: val})
					if key == "BREAKING CHANGE" || key == "BREAKING-CHANGE" {
						c.Breaking = true
					}
				} else if m := footerHashRe.FindStringSubmatch(line); m != nil {
					c.Footers = append(c.Footers, Footer{Key: m[1], Value: "#" + m[2]})
				}
			}
		}
		c.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}
	return c
}

// IsConventional returns true if Parse extracted a recognizable type.
func (c Commit) IsConventional() bool { return c.Type != "" }

// isFooterLine reports whether line is a valid CC footer token (either colon
// or hash-ref form).
func isFooterLine(line string) bool {
	return footerColonRe.MatchString(line) || footerHashRe.MatchString(line)
}

// findFooterStart returns the index within lines of the first line of the
// footer block, or -1 if no footer block is found.
//
// Per the CC spec, footers must be preceded by a blank line. We scan every
// blank-line boundary from the end and pick the first (from the end) position
// where all subsequent non-empty lines are valid footer tokens.
func findFooterStart(lines []string) int {
	// Collect blank-line boundary positions (index of the first line AFTER the blank).
	trimmed := make([]string, len(lines))
	for i, l := range lines {
		trimmed[i] = strings.TrimSpace(l)
	}
	var candidates []int
	for i := 1; i < len(trimmed); i++ {
		if trimmed[i-1] == "" && trimmed[i] != "" {
			candidates = append(candidates, i)
		}
	}
	// Try each boundary from the end.
	for j := len(candidates) - 1; j >= 0; j-- {
		start := candidates[j]
		ok := true
		for _, line := range lines[start:] {
			line = strings.TrimRight(line, " \t")
			if line == "" {
				continue
			}
			if !isFooterLine(line) {
				ok = false
				break
			}
		}
		if ok {
			return start
		}
	}
	return -1
}
