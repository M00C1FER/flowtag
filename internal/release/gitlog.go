// Package release wraps `git` invocations to read commit history.
package release

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/M00C1FER/flowtag/internal/conventional"
)

// LastTag returns the most recent annotated tag reachable from HEAD,
// or "" if no tag exists yet.
func LastTag() (string, error) {
	out, err := runGit("describe", "--tags", "--abbrev=0")
	if err != nil {
		// `git describe` exits non-zero when no tags exist; treat as empty.
		if strings.Contains(err.Error(), "No names found") || strings.Contains(strings.ToLower(out), "no names found") {
			return "", nil
		}
		// Some git versions print "fatal: No names found" on stderr.
		if strings.Contains(strings.ToLower(err.Error()), "fatal") {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// CommitsSince returns conventional commits from `since..HEAD`. If `since`
// is empty, it returns the entire history.
func CommitsSince(since string) ([]conventional.Commit, error) {
	args := []string{"log", "--no-merges", "--pretty=format:%H%x1f%s%x1f%b%x1e"}
	if since != "" {
		args = append(args, fmt.Sprintf("%s..HEAD", since))
	}
	out, err := runGit(args...)
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}
	records := strings.Split(out, "\x1e")
	commits := make([]conventional.Commit, 0, len(records))
	for _, rec := range records {
		rec = strings.Trim(rec, "\n ")
		if rec == "" {
			continue
		}
		fields := strings.SplitN(rec, "\x1f", 3)
		if len(fields) < 2 {
			continue
		}
		full := fields[1]
		if len(fields) == 3 && fields[2] != "" {
			full = fields[1] + "\n\n" + fields[2]
		}
		c := conventional.Parse(full)
		c.SHA = shortSHA(fields[0])
		commits = append(commits, c)
	}
	return commits, nil
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

func shortSHA(sha string) string {
	if len(sha) >= 7 {
		return sha[:7]
	}
	return sha
}
