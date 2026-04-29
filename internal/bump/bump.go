// Package bump computes the next semantic version from a list of conventional commits.
package bump

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/M00C1FER/flowtag/internal/conventional"
)

// Kind represents the size of a semver bump.
type Kind int

const (
	None Kind = iota
	Patch
	Minor
	Major
)

func (k Kind) String() string {
	switch k {
	case Major:
		return "major"
	case Minor:
		return "minor"
	case Patch:
		return "patch"
	default:
		return "none"
	}
}

// Rules maps conventional-commit types to a bump kind. Replaceable via YAML.
type Rules struct {
	// Type → bump kind.  Defaults: feat=minor, fix=patch, chore/docs/style=none, etc.
	TypeBump map[string]Kind
	// Default for unmatched types (Conventional commits with unknown `type:`).
	Default Kind
}

// DefaultRules is the conventional-commits standard mapping.
func DefaultRules() Rules {
	return Rules{
		TypeBump: map[string]Kind{
			"feat":     Minor,
			"fix":      Patch,
			"perf":     Patch,
			"refactor": None,
			"docs":     None,
			"style":    None,
			"chore":    None,
			"build":    None,
			"ci":       None,
			"test":     None,
			"revert":   Patch,
		},
		Default: None,
	}
}

// Compute reduces a list of commits to the largest required bump.
// Any breaking commit forces Major regardless of type.
func Compute(commits []conventional.Commit, rules Rules) Kind {
	highest := None
	for _, c := range commits {
		if !c.IsConventional() {
			continue
		}
		if c.Breaking {
			return Major // short-circuit
		}
		k, ok := rules.TypeBump[c.Type]
		if !ok {
			k = rules.Default
		}
		if k > highest {
			highest = k
		}
	}
	return highest
}

// Apply returns "major.minor.patch" advanced from `current` per the bump kind.
// `current` may have a leading "v"; the return preserves that prefix.
func Apply(current string, kind Kind) (string, error) {
	if kind == None {
		return current, nil
	}
	prefix := ""
	v := current
	if strings.HasPrefix(v, "v") {
		prefix = "v"
		v = v[1:]
	}
	if v == "" {
		v = "0.0.0"
	}
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("not a 3-part version: %q", current)
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	// Trim any pre-release tag from patch (e.g., 1.2.3-rc1 → 3)
	patchStr := parts[2]
	if dash := strings.IndexAny(patchStr, "-+"); dash >= 0 {
		patchStr = patchStr[:dash]
	}
	pat, err3 := strconv.Atoi(patchStr)
	if err1 != nil || err2 != nil || err3 != nil {
		return "", fmt.Errorf("non-numeric component in %q", current)
	}
	switch kind {
	case Major:
		maj++
		min = 0
		pat = 0
	case Minor:
		min++
		pat = 0
	case Patch:
		pat++
	}
	return fmt.Sprintf("%s%d.%d.%d", prefix, maj, min, pat), nil
}
