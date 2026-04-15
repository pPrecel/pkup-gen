# Claude Code Skills

This scenario shows how to use `pkup-gen` Claude Code skills to generate and enrich a PKUP report directly from a Claude Code conversation, without running any CLI commands manually.

## Prerequisites

Install the skills from the `pkup-gen` plugin marketplace:

```bash
claude plugin marketplace add pPrecel/pkup-gen
claude plugin install pkup-gen@pkup-gen
claude plugin install pkup-enchant@pkup-gen
```

You also need the `gh` CLI installed and authenticated:

```bash
gh auth status
```

## Available skills

| Skill | Trigger | Description |
|-------|---------|-------------|
| `pkup-gen` | `/pkup-gen` | Collects commits from GitHub and generates `.diff` files |
| `pkup-enchant` | `/pkup-enchant` | Groups commits into tasks, filters out non-creative work, writes Polish descriptions, and updates the report file |

The skills are designed to work together — `/pkup-gen` produces the raw output that `/pkup-enchant` then enriches. They can also be used independently.

---

## Scenario A: Full flow using both skills

The most common use case — run both skills in sequence within a single Claude Code session.

### Step 1 — Generate the report

Type in Claude Code:

```
/pkup-gen
```

The skill will:

1. Ask about the GitHub provider/org configuration and PKUP period
2. Verify `gh` CLI login for each provider
3. Detect your GitHub username and full name
4. Query the GitHub Search API for all your commits in the period
5. Download a `.diff` file for every commit into the output directory
6. Print a summary and suggest running `/pkup-enchant`

### Step 2 — Enrich the report

Type in Claude Code:

```
/pkup-enchant
```

The skill will:

1. Detect the output directory produced in Step 1
2. Fetch PR and issue context from GitHub for each commit
3. Group commits into coherent tasks (by linked issue, theme, or repository)
4. Show a classification table — propose which groups to include (creative work) and which to skip (version bumps, automated commits)
5. Ask for your approval; let you flip individual decisions
6. Delete `.diff` files for excluded groups
7. Write a one-sentence Polish description starting with _"Zaprojektowałem oraz zaimplementowałem..."_ for each included group
8. Overwrite the report file (`.txt` or `.docx`) with the enriched content

---

## Scenario B: CLI generation + skill enrichment

Use this when you already generated the report with the `pkup-gen` CLI binary and want to enrich it without repeating the collection step.

### Step 1 — Generate the report using the CLI

```bash
pkup gen \
  --username pPrecel \
  --org kyma-project \
  --output reports/FILIP_STROZIK
```

This produces a directory like `reports/FILIP_STROZIK/` containing `.diff` files and a `report.txt`.

For a more complex setup with multiple providers or a `.docx` template, use `pkup compose`:

```bash
pkup compose --config .pkupcompose.yaml
```

See [compose-and-send](../compose-and-send/README.md) for the full config reference.

### Step 2 — Enrich with the skill

Open Claude Code in the repo directory and type:

```
/pkup-enchant
```

The skill detects the existing output directory and picks up from there — no re-collection needed.

---

## Combining skills with other scenarios

The skills work on any directory that contains `.diff` files produced by `pkup-gen` — regardless of how those files were generated. This means you can mix and match:

- Generate via CLI → enrich via skill
- Generate via skill → inspect diffs manually → enrich via skill
- Generate via `pkup compose` for multiple users → run `/pkup-enchant` separately for each output directory
