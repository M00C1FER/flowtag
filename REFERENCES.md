# Reference projects studied

Reviewed during the 2026-05-02 audit cycle to identify patterns adopted by
high-star peers in the Conventional Commits / semver-bump space.

| Repo | Stars | License | Key pattern adopted |
|---|--:|---|---|
| [orhun/git-cliff](https://github.com/orhun/git-cliff) | ~10 k | MIT/Apache-2 | Footer section starts after last blank-line boundary (§ spec-correct body/footer split); hash-ref trailer form (`token #value`) is an explicit grammar rule |
| [googleapis/release-please](https://github.com/googleapis/release-please) | ~6 k | Apache-2 | `BREAKING-CHANGE:` (hyphen) treated as equivalent to `BREAKING CHANGE:` (space) — both RFC-8288-style and human-style tokens must fire the major bump |
| [semantic-release/semantic-release](https://github.com/semantic-release/semantic-release) | ~21 k | MIT | Bump algorithm short-circuits on first breaking commit; does **not** continue scanning remaining commits — matches flowtag `Compute()` behaviour |
| [conventional-changelog/conventional-changelog](https://github.com/conventional-changelog/conventional-changelog) | ~7 k | ISC | CHANGELOG sections ordered: Breaking → Features → Fixes → Perf → Refactors → Docs → Tests → Maintenance (flowtag mirrors this ordering) |
| [commitizen-tools/commitizen](https://github.com/commitizen-tools/commitizen) | ~5 k | MIT | Cross-platform installer detects `TERMUX_VERSION` env var to select `pkg` as the package manager — adopted in `install.sh` |

## Specific patterns incorporated

### 1. Body / footer separation (git-cliff / CC spec §12)

Footers must follow an **explicit blank line** after the body.  The original
flowtag parser scanned every post-subject line for footer patterns, which could
misclassify body text like `"See Doc: section above"` as a footer.  Fixed by
`findFooterStart()` in `internal/conventional/parser.go`.

### 2. Hash-ref footer form (git-cliff)

The CC spec allows `token #value` as an alternative to `token: value` (e.g.
`Closes #101`).  Added `footerHashRe` and corresponding parsing logic.

### 3. `BREAKING-CHANGE:` equivalence (release-please)

The hyphen form (`BREAKING-CHANGE:`) is listed in the CC spec as an allowed
alternative to `BREAKING CHANGE:`.  Ensured `c.Breaking = true` fires for
both variants.

### 4. Termux support (commitizen)

`install.sh` now detects Termux via `TERMUX_VERSION` and routes to `pkg
install` instead of `apt`/`dnf`/`pacman`.

## Known gaps vs. peers

- **Multi-line footer continuations** (a footer value spanning multiple lines)
  are not yet parsed.  git-cliff supports them; tracked as a follow-up.
- **YAML bump-rule overrides** (`--rules`) are a v0.2 item; semantic-release
  and git-cliff both support config-file-driven type→bump mappings.
