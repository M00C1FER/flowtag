// flowtag — conventional-commits → semver bump + CHANGELOG + draft release.
//
// Usage:
//
//	flowtag --next-version           # print the bumped version (or current if no bump needed)
//	flowtag --bump                   # print the bump kind: major | minor | patch | none
//	flowtag --changelog              # render markdown to stdout
//	flowtag --write-changelog        # prepend CHANGELOG.md with the new entry
//	flowtag --dry-run --release      # print what a release would do
//	flowtag --rules my-rules.yaml    # override the bump-type mapping
//
// The release publisher (tagging + GitHub Release draft) is intentionally
// kept out of v0.1 — see ROADMAP.md. v0.1 is read-only + write-changelog.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/M00C1FER/flowtag/internal/bump"
	"github.com/M00C1FER/flowtag/internal/changelog"
	"github.com/M00C1FER/flowtag/internal/release"
)

func main() {
	var (
		nextVersion     = flag.Bool("next-version", false, "print bumped version and exit")
		showBump        = flag.Bool("bump", false, "print bump kind (major|minor|patch|none)")
		printChangelog  = flag.Bool("changelog", false, "print CHANGELOG markdown to stdout")
		writeChangelog  = flag.Bool("write-changelog", false, "prepend CHANGELOG.md with new entry")
		rulesFile       = flag.String("rules", "", "YAML file with custom bump rules (optional)")
		sinceOverride   = flag.String("since", "", "override 'last tag' (default: auto-detect via git describe)")
		baseVersion     = flag.String("base", "", "override current version (default: auto-detect last tag)")
	)
	flag.Parse()

	if !(*nextVersion || *showBump || *printChangelog || *writeChangelog) {
		flag.Usage()
		os.Exit(2)
	}

	rules := bump.DefaultRules()
	if *rulesFile != "" {
		// v0.1 — YAML rule loading is a follow-up. For now error if user
		// passes --rules so they can't get false-positive runs.
		fmt.Fprintln(os.Stderr, "[flowtag] --rules YAML loading not yet implemented in v0.1")
		os.Exit(2)
	}

	since := *sinceOverride
	current := *baseVersion
	if since == "" || current == "" {
		tag, err := release.LastTag()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[flowtag] git describe failed: %v\n", err)
			os.Exit(1)
		}
		if since == "" {
			since = tag
		}
		if current == "" {
			current = tag
		}
	}

	commits, err := release.CommitsSince(since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[flowtag] git log failed: %v\n", err)
		os.Exit(1)
	}

	kind := bump.Compute(commits, rules)
	if *showBump {
		fmt.Println(kind)
		return
	}

	next, err := bump.Apply(current, kind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[flowtag] cannot bump %q: %v\n", current, err)
		os.Exit(1)
	}
	if *nextVersion {
		fmt.Println(next)
		return
	}

	when := time.Now().UTC()
	rendered := changelog.Render(next, when, commits)

	if *printChangelog {
		fmt.Print(rendered)
		return
	}
	if *writeChangelog {
		if err := prependChangelog("CHANGELOG.md", rendered); err != nil {
			fmt.Fprintf(os.Stderr, "[flowtag] write CHANGELOG.md: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "[flowtag] prepended %s entry to CHANGELOG.md\n", next)
	}
}

// prependChangelog inserts `entry` at the top of the file, creating it if
// missing. The Keep-a-Changelog convention places newest at top.
func prependChangelog(path, entry string) error {
	existing, _ := os.ReadFile(path) // missing-file → empty
	header := "# Changelog\n\n"
	body := string(existing)
	body = strings.TrimPrefix(body, header)
	combined := header + entry + "\n" + body
	return os.WriteFile(path, []byte(combined), 0644)
}
