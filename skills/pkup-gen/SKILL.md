---
name: pkup-gen
description: Generate a PKUP report for given GitHub users. Use when the user asks to generate a PKUP report.
---

Generate a PKUP report.

## What is PKUP

PKUP (Podwyższone Koszty Uzyskania Przychodu) is a Polish tax concept. A PKUP report documents creative programming work done during a given month. It contains a list of commits/PRs as evidence of creative work, along with `.diff` files.

## Step 0: Inform about planned operations and get consent

Before asking any configuration questions, **detect the current permission mode** by checking the environment: if the permission mode is not **Auto**, display a warning and ask for confirmation before proceeding. Then display the information block.

### 0a. Detect permission mode

Check whether Claude Code is running in Auto permission mode (i.e. tool calls are approved automatically without prompting the user each time).

If **not** in Auto mode — use `AskUserQuestion` to warn the user:

> **Warning:** This skill will make many bash, `gh`, `curl`, and `jq` calls. In the current permission mode you will be prompted to approve **each one individually**, which may result in dozens of confirmation dialogs.
>
> You can switch to Auto mode now with `Shift+Tab` or the toolbar icon to avoid this.
>
> Do you want to continue anyway?

Options:
- "Yes, continue anyway"
- "No, I'll switch to Auto mode first and re-run"
- "No, cancel"

If the user selects "No, I'll switch to Auto mode first and re-run" or "No, cancel" — stop immediately.

If the user confirms or Auto mode is already active — display the information block below and continue.

---

**Generating a PKUP report requires the following operations:**

**Programs and tools:**
- `bash` — executing shell commands
- `gh` (GitHub CLI) — GitHub API queries (searching commits, fetching PRs and issues, checking login status)
- `curl` — downloading `.diff` files from the GitHub API
- `jq` — processing JSON responses
- `mkdir` — creating the output directory

**Network operations:**
- Requests to `api.github.com` (GitHub.com Search API, commits, pulls, issues)
- Requests to `github.tools.sap/api/v3` (GitHub Enterprise API) — if applicable

**Local file operations:**
- Creating the output directory (e.g. `reports/pPrecel/`)
- Writing `.diff` files for each qualified commit
- Writing the `report_enchanted.txt` file

---

Calculate the default PKUP period: from the 19th of the previous month (00:00:00) to the 18th of the current month (23:59:59).

Ask the user (AskUserQuestion) **three questions at once**:

**Question 1 — consent:**
> May I proceed with generating the report using the tools and permissions listed above?

Options:
- "Yes, proceed"
- "No, cancel"

**Question 2 — provider and organization configuration:**
> Select the GitHub provider and organization configuration:

Options:
- "Default: github.com → kyma-project, github.tools.sap → kyma"
- "Custom configuration"

**Question 3 — PKUP period:**
> Calculated PKUP period: **DD.MM.YYYY – DD.MM.YYYY**. Is this period correct?

Options:
- "Yes, use this period"
- "No, I will provide a custom period"

Ask all three questions at once in a single AskUserQuestion call (three entries in the `questions` array).

If the user refuses in question 1 — stop.
If the user selects "No, I will provide a custom period" — ask for dates in `DD.MM.YYYY - DD.MM.YYYY` format before proceeding.

## Step 1: Determine Git providers and organizations

Based on the answers from Step 0:

**If "Default" was selected:**
- **Providers:** `github.com`, `github.tools.sap`
- **Organizations:** `kyma-project` (github.com), `kyma` (github.tools.sap)

**If "Custom configuration" was selected:**
Ask the user for details:
- List of providers (e.g. `github.com`, `github.tools.sap`, or another address)
- For each provider: list of organizations or repositories (format `org:NAME` or `repo:ORG/REPO`)

Remember the established values — they will be used in subsequent steps.

## Step 2: Check gh CLI login status

Run `gh auth status 2>&1` and check the result for each selected provider.

For each provider one of the following cases may occur:

**Logged in successfully** (`✓ Logged in`) — continue.

**Not logged in or token invalid** — display the instruction:

```
You are not logged in to <PROVIDER>. Run in the terminal:

  ! gh auth login -h <PROVIDER>

After logging in, return to this conversation.
```

Wait for the user to confirm they have logged in, then re-check `gh auth status` and make sure the login succeeded.

Only proceed to the next step once all selected providers have an active, valid session.

## Step 3: Determine usernames

For each selected provider, determine the username:

- Read the logged-in user from the `gh auth status` output (the `account <USERNAME>` line).
- If it cannot be determined automatically — ask the user.

Also determine:
- **outputDir** — the output directory, e.g. `reports/FIRST_LAST` (ask if unknown)

## Step 4: Collect commits

For each user and each org/repo, query the GitHub Search API:

```bash
gh api --method=GET search/commits \
  --hostname PROVIDER_HOSTNAME \
  -f q="author:USERNAME org:ORG committer-date:>=YYYY-MM-DD committer-date:<=YYYY-MM-DD" \
  -f per_page=100 \
  --paginate \
  | jq -s '[.[] | .items[] | {
      sha: .sha,
      repo: .repository.full_name,
      message: (.commit.message | split("\n")[0]),
      date: .commit.author.date
    }]
    | unique_by(.sha)
    | map(select(.date >= "SINCE_DATE" and .date <= "UNTIL_DATE"))
    | sort_by(.date)'
```

For a repo instead of an org: use `repo:ORG/REPO` instead of `org:ORG`.

For `github.com` use `--hostname github.com` (or omit it, as it is the default host).
For `github.tools.sap` use `--hostname github.tools.sap`.

**Important:** The GitHub Search API does not filter dates server-side — always filter client-side (`select(.date >= ...)`). Use `unique_by(.sha)` to avoid duplicates from different branches.

## Step 5: Filter — creative work only

Based on PR titles and bodies, evaluate each commit against the criteria below, then **ask the user** to approve the classification using `AskUserQuestion` with `multiSelect: true`.

**Evaluation criteria:**

Pre-**uncheck** (proposed exclusion):
- Version bumps for libraries, dependencies, container images (bump deps, upgrade X to Y, bump X version)
- Automated commits (dependabot, renovate, sync/retrigger)
- Updates to developer tooling only (linter version, go version, actions versions)

Pre-**check** (proposed inclusion):
- Implementing new features or systems
- Linked to GitHub issues
- Creating or updating technical documentation
- Describing architectural or design decisions
- Non-trivial bug fixes requiring analysis

When in doubt, check the PR: `gh api --hostname PROVIDER_HOSTNAME repos/ORG/REPO/pulls/PR_NUMBER`

**Presentation to the user:**

Display a markdown table with the classification of all commits:

```
| # | Repo | PR | Title | Date | Decision | Reason |
|---|------|----|-------|------|----------|--------|
| 1 | kyma-project/serverless | #2400 | Performance test hello_world scenario | 27.03 | ✅ INCLUDE | linked to issue #2082 |
| 2 | kyma-project/keda-manager | #823 | Upgrade FIPS keda version | 19.03 | ❌ SKIP | version bump |
```

Then call `AskUserQuestion` with the question:

> Do you want to keep this classification or make changes?

Options:
- "Keep this classification and continue"
- "I want to change something"

If the user selects "I want to change something" — ask them in plain text to provide the numbers of commits whose decision they want to flip (e.g. "change 2, 5, 7"), apply the changes, display the updated table, and ask for approval again.

After the user approves — continue only with commits marked ✅ INCLUDE.

## Step 6: Download .diff files (creative work only)

For each **qualified** commit, download the diff and save it to `outputDir`:

```bash
curl -sf \
  -H "Authorization: token $(gh auth token --hostname PROVIDER_HOSTNAME)" \
  -H "Accept: application/vnd.github.v3.diff" \
  "https://PROVIDER_HOSTNAME/repos/ORG/REPO/commits/SHA" \
  > "outputDir/ORG_REPO_SHA8.diff"
```

For `github.com` use `https://api.github.com/repos/...`.
For `github.tools.sap` use `https://github.tools.sap/api/v3/repos/...`.

File name format: `{org}_{repo}_{first_8_chars_of_sha}.diff`

Do not download diffs for commits excluded as non-creative work.

## Step 7: Generate report_enchanted.txt

This is the only report file — an enriched version with descriptions of creative work written in Polish (required by Polish tax regulations).

### 7a. Gather context from GitHub

For each qualified commit, fetch the associated PR and check for linked issues:

```bash
gh api --hostname PROVIDER_HOSTNAME repos/ORG/REPO/pulls/PR_NUMBER \
  | jq -r '.body' | grep -E "issues/[0-9]+"
```

For found issues, fetch their titles and descriptions:

```bash
gh api --hostname PROVIDER_HOSTNAME repos/ORG/REPO/issues/ISSUE_NUMBER \
  | jq '{title, body: .body[0:500]}'
```

### 7b. Group commits

Group commits by linked issue (if several PRs point to the same issue — they form one group). Commits without issues should be grouped by repository or thematically if a common goal is apparent.

### 7c. Write task descriptions

For each group write **one description** starting with:
> Zaprojektowałem oraz zaimplementowałem ... *(I designed and implemented ...)*

Rules:
- **Do not mention mechanical actions** (version changes, image updates)
- **Emphasize the creative and inventive nature** of the work
- **Describe business or technical value** — what it enables, what problem it solves
- **Use language that indicates authorship** — "zaprojektowałem" (I designed), "stworzyłem" (I created), "opracowałem" (I developed)
- The description must be a single sentence, written in Polish

### 7d. File format

```
period:
DD.MM.YYYY - DD.MM.YYYY

approvalDate:
DD.MM.YYYY

result:

--- org/repo ---

Zaprojektowałem oraz zaimplementowałem ...
  - PR Title (#NR) (org_repo_sha8.diff)
  - PR Title (#NR) (org_repo_sha8.diff)

--- org/repo ---

Zaprojektowałem oraz zaimplementowałem ...
  - PR Title (#NR) (org_repo_sha8.diff)
```

`approvalDate` = the day after the end of the period (the 19th of the current month).

## Examples of good descriptions

- "Zaprojektowałem oraz zaimplementowałem kompleksową platformę do pomiaru wydajności funkcji serverless, umożliwiającą deweloperom świadome podejmowanie decyzji architektonicznych poprzez dostarczenie danych o opóźnieniach platformy dla różnych środowisk uruchomieniowych i profili zasobów."
- "Zaprojektowałem oraz zaimplementowałem automatyczny mechanizm budowania i dystrybucji bezpiecznych, zgodnych ze standardem FIPS obrazów kontenerowych dla KEDA, zapewniający powtarzalne i audytowalne środowisko uruchomieniowe spełniające wymagania bezpieczeństwa środowisk korporacyjnych."
- "Zaprojektowałem oraz zaimplementowałem izolację modułu Docker Registry we własnej dedykowanej przestrzeni nazw Kubernetes, poprawiając separację zasobów i umożliwiając niezależne zarządzanie cyklem życia modułu w klastrze."

## Technical notes

- `gh api --paginate` handles pagination automatically
- GitHub Search API has a limit of 1000 results per query
- Rate limit: 5000 req/h for PAT tokens
- The output `outputDir` should contain only: `report_enchanted.txt` + `.diff` files for creative work
