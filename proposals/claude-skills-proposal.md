# Proposal: Claude Code Skills for pkup-gen

## Summary

Add two Claude Code skills — `/pkup-gen` and `/pkup-enchant` — that let users generate and enrich PKUP reports directly from a Claude Code conversation, without installing or running the `pkup-gen` binary.

## Motivation

The `pkup-gen` CLI requires installation, configuration, and familiarity with its flags and YAML config format. Users who already work in Claude Code (the AI coding assistant) can instead invoke skills that guide them through the process interactively, handle edge cases conversationally, and produce a fully enriched report without ever leaving the chat.

This also unlocks a hybrid scenario: users who prefer the CLI for report generation can still use `/pkup-enchant` to automate the enrichment step (grouping, filtering, Polish descriptions) that the CLI does not provide.

## Skills

### `/pkup-gen`

Collects commits from one or more GitHub providers and downloads `.diff` files for the PKUP period.

**Flow:**
1. Ask the user about provider/org configuration and confirm the PKUP period (19th of previous month → 18th of current month)
2. Verify `gh` CLI login for each provider; prompt the user to log in if needed
3. Resolve the GitHub username and full name for each provider
4. Query the GitHub Search API (`search/commits`) for all commits in the period, using both `author:LOGIN` and `author-name:"Full Name"` to catch all signatures; deduplicate by SHA
5. Download a `.diff` file per commit via the GitHub REST API and save to the output directory
6. Print a summary and suggest running `/pkup-enchant`
7. Ask the user whether to star the `pPrecel/pkup-gen` repository

**Output:** A directory containing `.diff` files named `{org}_{repo}_{sha8}.diff`.

### `/pkup-enchant`

Enriches an existing report directory — groups commits into tasks, filters out non-creative work, writes Polish descriptions, and updates the report file.

**Flow:**
1. Detect the output directory (list `reports/` subdirectories or ask for a custom path)
2. Determine required GitHub providers from diff filenames; verify `gh` CLI login
3. Fetch PR and linked issue context from GitHub for every commit
4. Group commits into coherent tasks by linked issue, theme, or repository
5. Show a numbered classification table with proposed include/skip decisions; ask for approval
6. Allow the user to flip individual decisions; redisplay the table after each change
7. Delete `.diff` files for excluded groups
8. Write a one-sentence Polish description per included group (_"Zaprojektowałem oraz zaimplementowałem..."_)
9. Overwrite the report file (`.txt` or `.docx`) with the enriched grouped result
10. Ask the user whether to star the `pPrecel/pkup-gen` repository

**Input:** A directory with `.diff` files and a report file produced by `/pkup-gen` or `pkup-gen` CLI.

## Compatibility

Both skills operate on the same directory format produced by the existing `pkup-gen` binary and `pkup compose` command. No changes to the core Go codebase are required.

## Distribution

Skills are distributed as a Claude Code plugin marketplace at `pPrecel/pkup-gen`. Users install with:

```bash
claude plugin marketplace add pPrecel/pkup-gen
claude plugin install pkup-tools@pkup-gen
```

Plugin metadata lives in `.claude-plugin/marketplace.json` and `.claude-plugin/manifest.json`.

## Dependencies

- `gh` CLI — authenticated to each required GitHub provider
- `curl`, `jq` — used internally by the skills for API calls and JSON processing
- `python3` — used by `/pkup-enchant` when patching `.docx` report files
