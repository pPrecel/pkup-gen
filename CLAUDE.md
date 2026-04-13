# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build      # builds binary to .out/pkup
make verify     # runs go mod verify, gofmt, go vet, and tests with race detection
go test --cover --race -timeout 30s ./...          # run all tests
go test --cover --race -timeout 30s ./pkg/github/  # run tests in a single package
```

## Architecture

`pkup-gen` generates PKUP (Polish tax deduction) reports documenting creative GitHub contributions for a monthly period (19th of previous month → 18th of current month).

**CLI entry points** (`cmd/`):
- `gen` — single-user report generation via CLI flags
- `compose` — multi-user/multi-org automation driven by a YAML config (`.pkupcompose.yaml`)
- `send` — emails zipped reports via SMTP
- `version` — prints build metadata

**Data flow for `compose`:**
1. `pkg/config` — parses YAML config (remotes, orgs, repos, users, email settings)
2. `pkg/compose/utils/clients.go` — lazily creates one `pkg/github` client per enterprise URL
3. `pkg/compose/utils/authors.go` — resolves GitHub usernames → commit author signatures
4. `pkg/compose/utils/commits.go` — `LazyCommitsLister`: fetches and caches commits per repo/org, deduplicates across branches
5. `pkg/artifacts` — writes `.diff` files to the output directory
6. `pkg/report` — renders `report.txt` or a `.docx` from a template (keyword substitution)
7. `pkg/send` — zips the output directory and sends it via SMTP

**GitHub client** (`pkg/github/`):
- Supports both github.com and GitHub Enterprise (configurable hostname)
- Automatic rate-limit retry (up to 5 attempts with backoff) applied to every API call
- `commit.go` filters by time window and deduplicates SHAs across branches
- `diff.go` fetches raw diff content per commit

**Auth** (`internal/token/`): OAuth flow via the `pkup-gen` GitHub App; token cached in the OS keyring. PAT can be passed via `--token` flag instead.

**Terminal UI** (`internal/view/`): pterm-based multi-task progress display used during concurrent report generation.

**Concurrency**: `compose` processes users in parallel goroutines; `LazyCommitsLister` uses a mutex to ensure each repo/org is fetched only once even under concurrent access.

## Configuration

`.pkupcompose.yaml` is the compose config format. Fields: `remotes` (enterprise URLs + tokens), `reports` (users, orgs, repos, output dir, template path), `send` (SMTP settings). See `.pkupcompose.yaml` in the repo root for a full annotated example.

## Release

GoReleaser builds cross-platform binaries (Linux/macOS, multiple arches) and publishes to the `pPrecel/homebrew-tap` Homebrew formula on git tag push.
