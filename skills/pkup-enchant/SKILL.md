---
name: pkup-enchant
description: Enrich an existing PKUP report directory — group commits into tasks, filter out non-creative groups, and write Polish descriptions. Use when the user has already generated a PKUP report and wants to clean it up or add "Zaprojektowałem..." descriptions.
---

Enrich an existing PKUP report.

## What this skill does

Starting from an already-generated output directory (containing `.diff` files and a report file), this skill:
1. Determines which GitHub providers are needed based on diff filenames
2. Verifies `gh` CLI login status for those providers
3. Fetches PR/issue context from GitHub for all commits
4. Groups commits into tasks (by linked issue, theme, or repository)
5. Presents the grouped tasks for classification and lets you approve which groups to keep
6. Removes excluded diffs and their entries from the report
7. Writes one-sentence Polish descriptions per group
8. Applies the enriched result block directly to the original report file (`.txt` or `.docx`)

## Step 0: Determine the report directory

List subdirectories of `reports/` (if it exists) and ask the user which directory to process:

```bash
ls reports/
```

Ask (AskUserQuestion) with the found directories as options plus "Custom path" as a fallback. The selected directory must contain at least one `.diff` file and one report file (`*.txt` or `*.docx`).

Remember the path as **outputDir**.

## Step 1: Determine required providers and verify login

Scan all `.diff` filenames in **outputDir** and derive required providers:

- Filename starting with `kyma-project_` → requires **github.com**
- Filename starting with `kyma_` → requires **github.tools.sap**
- Any other prefix → ask the user which provider hosts that repository

Run:

```bash
gh auth status 2>&1
```

For each required provider:

**Logged in** (`✓ Logged in to <PROVIDER>`) — continue.

**Not logged in or token invalid** — display:

```
Nie jesteś zalogowany do <PROVIDER>. Uruchom w terminalu:

  ! gh auth login -h <PROVIDER>

Po zalogowaniu wróć do tej rozmowy.
```

**Stop** — do not continue until all required providers have a valid session. After the user confirms they logged in, re-run `gh auth status` to verify.

## Step 2: Fetch GitHub context for all commits

Read every `.diff` filename in **outputDir**. For each file, parse:
- `ORG` and `REPO` — from the prefix (e.g. `kyma-project_serverless_74ee55ce.diff` → org=`kyma-project`, repo=`serverless`)
- `SHA8` — the last segment before `.diff` (e.g. `74ee55ce`)

Determine the hostname:
- `kyma-project_*` → `github.com` (omit `--hostname`)
- `kyma_*` → `github.tools.sap`

For each commit, find the associated PR:

```bash
gh api --method=GET search/issues \
  --hostname PROVIDER \
  -f q="repo:ORG/REPO SHA8 type:pr" \
  --jq '.items[0] | {number: .number, title: .title, body: (.body // "" | .[0:600])}'
```

For found PRs, check for linked issues in the body (patterns: `issues/NUMBER`, `Resolves #N`, `Fixes #N`).

For found issues, fetch title and description:

```bash
gh api --hostname PROVIDER repos/ORG/REPO/issues/ISSUE_NUMBER \
  --jq '{title, body: (.body // "" | .[0:500])}'
```

Build a mapping: `diff_filename → {pr_number, pr_title, issue_number, issue_title, issue_body}`.

## Step 3: Group commits into tasks

Group all diffs into tasks using the following priority:

1. **Same linked issue** → one group (even across multiple repos)
2. **Same repo, no common issue, similar theme** → one group
3. **Otherwise** → one group per repository

Each group represents one task. A group may contain any number of commits — a single small commit ("fix date", "fix format") is a valid group if it belongs to a coherent task.

## Step 4: Classify groups and ask for approval

For each group, propose a classification:

**❌ SKIP — proposed exclusion** (apply only when the entire group, viewed as a whole, is clearly non-creative):
- The group consists solely of library/dependency/image version bumps
- The group consists solely of automated commits (dependabot, renovate)
- The group consists solely of single-line config changes with no analytical background

**✅ INCLUDE — proposed inclusion** (default for everything else):
- Any group implementing a feature, fixing a non-trivial bug, or making architectural decisions
- Any group linked to a GitHub issue
- Any group containing documentation or design work
- **Any small commit that is part of a larger implementation chain** — even if the commit message looks trivial ("fix date", "fix format", "cleanup"), include it if the PR or issue context suggests creative work

**Important:** Evaluate groups, not individual commits. A small commit inside a creative task is creative by association. When in doubt — include.

Display a numbered classification table grouped by task:

```
| # | Task (issue / theme) | Commits | Decision | Reason |
|---|----------------------|---------|----------|--------|
| 1 | serverless#2082 — Design load test for function serving layer | 10 commits | ✅ INCLUDE | new feature, linked to issue |
| 2 | keda-manager — Rebuild FIPS images from Chainguard | 3 commits | ✅ INCLUDE | new workflow, linked to issue |
| 3 | keda-manager — Upgrade FIPS keda to 1.19.0 | 1 commit  | ❌ SKIP   | version bump only |
```

The `#` column must always be present and contain sequential numbers starting from 1. Never omit the table or the numbers.

Then ask (AskUserQuestion):

> Czy chcesz zachować tę klasyfikację, czy wprowadzić zmiany?

Options:
- "Zatwierdź klasyfikację i kontynuuj"
- "Chcę coś zmienić"

If the user selects "Chcę coś zmienić" — ask in plain text which numbers to flip (e.g. "change 2, 5"). Apply the changes, then **always redisplay the full numbered table** with the updated decisions before asking for approval again. Repeat this loop until the user approves.

After approval — proceed only with groups marked ✅ INCLUDE.

## Step 5: Remove excluded diffs

For each group marked ❌ SKIP, delete all its `.diff` files from **outputDir**:

```bash
rm outputDir/filename.diff
```

Print a summary of deleted files.

## Step 6: Write Polish descriptions

For each group marked ✅ INCLUDE, write **one description sentence** starting with:
> Zaprojektowałem oraz zaimplementowałem ...

Rules:
- Do **not** mention mechanical actions (version numbers, image updates)
- Emphasise the **creative and inventive nature** of the work
- Describe the **business or technical value** — what it enables, what problem it solves
- Use language that indicates **authorship**: "zaprojektowałem" (I designed), "stworzyłem" (I created), "opracowałem" (I developed)
- Single sentence, written in **Polish**

Good description examples:
- "Zaprojektowałem oraz zaimplementowałem kompleksową platformę do pomiaru wydajności funkcji serverless, umożliwiającą deweloperom świadome podejmowanie decyzji architektonicznych poprzez dostarczenie danych o opóźnieniach platformy dla różnych środowisk uruchomieniowych i profili zasobów."
- "Zaprojektowałem oraz zaimplementowałem automatyczny mechanizm budowania i dystrybucji bezpiecznych, zgodnych ze standardem FIPS obrazów kontenerowych dla KEDA, zapewniający powtarzalne i audytowalne środowisko uruchomieniowe spełniające wymagania bezpieczeństwa środowisk korporacyjnych."
- "Zaprojektowałem oraz zaimplementowałem dokumentację techniczną dostępnych akcji GitHub Actions, umożliwiającą deweloperom samodzielne i poprawne korzystanie z udostępnianych narzędzi automatyzacji."

## Step 7: Apply the result to the report file

The `result` block has no repo separators — it is a flat sequence of description+commits groups, separated by **3 blank lines** between groups:

```
Zaprojektowałem oraz zaimplementowałem ...
  - PR Title (#NR) (org_repo_sha8.diff)
  - PR Title (#NR) (org_repo_sha8.diff)



Zaprojektowałem oraz zaimplementowałem ...
  - PR Title (#NR) (org_repo_sha8.diff)
```

Read the report file — the first `*.txt` or `*.docx` file found in **outputDir** (skip `report_enchanted.txt` if present — it is a legacy file).

Apply the result differently depending on the report file type:

### `.txt` report

Preserve the `period:` and `approvalDate:` values from the original. Overwrite the entire file with:

```
period:
DD.MM.YYYY - DD.MM.YYYY

approvalDate:
DD.MM.YYYY

result:

<result block>
```

`approvalDate` — keep from the original if present; otherwise set to the day after the period end date.

### `.docx` report

The `.docx` is a ZIP archive containing `word/document.xml`. Edit it directly with Python using `zipfile`.

**Locate the commit list region** — the document already contains a flat list of commit entries written by pkup-gen. Find the XML `<w:p>` paragraph that contains the first commit entry (the first `.diff` filename visible in the document) and the paragraph containing the last commit entry. Replace that entire region with the grouped structure.

**Paragraph types to generate:**

1. **Description paragraph** (plain text, no indent) — matches the style of plain body text in the document (no `<w:pStyle>`, no `<w:numPr>`):
```xml
<w:p w14:paraId="RAND8HEX" w14:textId="77777777" w:rsidR="00D64311" w:rsidRDefault="00D64311">
  <w:r>
    <w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>
    <w:t xml:space="preserve">Zaprojektowałem oraz zaimplementowałem ...</w:t>
  </w:r>
</w:p>
```

2. **Commit entry paragraph** (bulleted list, indented) — copy the `<w:pPr>` with `<w:pStyle w:val="ListParagraph"/>` and `<w:numPr>` from the original commit paragraphs found in the document.

3. **Empty separator paragraph** — same structure as the description paragraph but with no `<w:r>` content. Insert one between groups (not before the first group).

**Python script template:**

```bash
python3 - <<'EOF'
import zipfile, os, random

src = "outputDir/report.docx"
tmp = src + ".tmp"

def rand_id():
    return f"{random.randint(0x10000000, 0xEFFFFFFF):08X}"

def make_empty_para():
    return (
        f'<w:p w14:paraId="{rand_id()}" w14:textId="77777777" w:rsidR="00D64311" w:rsidRDefault="00D64311">'
        f'<w:pPr><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr></w:p>'
    )

def make_desc_para(text):
    return (
        f'<w:p w14:paraId="{rand_id()}" w14:textId="77777777" w:rsidR="00D64311" w:rsidRDefault="00D64311">'
        f'<w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>'
        f'<w:t xml:space="preserve">{text}</w:t></w:r></w:p>'
    )

def make_commit_para(text):
    # Copy <w:pPr> with pStyle/numPr from original commit paragraphs in the document
    return (
        f'<w:p w14:paraId="{rand_id()}" w14:textId="77777777" w:rsidR="00D64311" w:rsidRDefault="00D64311" w:rsidP="00DF69A1">'
        f'<w:pPr><w:pStyle w:val="ListParagraph"/><w:numPr><w:ilvl w:val="1"/><w:numId w:val="8"/></w:numPr>'
        f'<w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr>'
        f'<w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>'
        f'<w:t xml:space="preserve">{text}</w:t></w:r></w:p>'
    )

groups = [
    # { "desc": "Zaprojektowałem ...", "commits": ["PR Title (#N) (file.diff)", ...] },
]

with zipfile.ZipFile(src, 'r') as z:
    with z.open('word/document.xml') as f:
        content = f.read().decode('utf-8')

# Build new XML
new_xml = ""
for i, g in enumerate(groups):
    if i > 0:
        new_xml += make_empty_para()
    new_xml += make_desc_para(g["desc"])
    for c in g["commits"]:
        new_xml += make_commit_para(c)

# Find region: from <w:p> containing first commit to </w:p> of last commit
first_text = "FIRST_COMMIT_TITLE"   # text of first commit entry in the document
last_text  = "LAST_COMMIT_TITLE"    # text of last commit entry in the document

idx_first = content.find(first_text)
idx_last  = content.find(last_text)
para_start = content.rfind('<w:p ', 0, idx_first)
para_end   = content.find('</w:p>', idx_last) + len('</w:p>')

new_content = content[:para_start] + new_xml + content[para_end:]

with zipfile.ZipFile(src, 'r') as zin, zipfile.ZipFile(tmp, 'w', zipfile.ZIP_DEFLATED) as zout:
    for item in zin.infolist():
        data = zin.read(item.filename)
        if item.filename == 'word/document.xml':
            data = new_content.encode('utf-8')
        zout.writestr(item, data)

os.replace(tmp, src)
print("Done")
EOF
```

**Important:** Before running the script, inspect the actual `word/document.xml` to find:
- The exact text of the first and last commit entries (use as `first_text` / `last_text`)
- The `<w:pPr>` structure of original commit paragraphs (to copy correct `numId` and `ilvl` values into `make_commit_para`)

Also replace `pkupGenPeriodFrom`, `pkupGenPeriodTill`, and `pkupGenApprovalDate` placeholders if they still contain the template strings (i.e. were not yet filled in by pkup-gen).

After writing the file, print the full result block to the conversation so the user can review it.

## Step 8: Thank the user and ask about starring the repo

Print the following message:

```
Dzięki za skorzystanie z pkup-gen! 🙌
Jeśli narzędzie Ci pomogło, rozważ oznaczenie gwiazdką na GitHubie —
to najlepszy sposób, żeby docenić projekt i pomóc innym go odkryć.

  ⭐ https://github.com/pPrecel/pkup-gen
```

Then ask (AskUserQuestion):

> Czy chcesz, żebym oznaczył gwiazdką repozytorium pPrecel/pkup-gen za Ciebie?

Options:
- "Tak, oznacz gwiazdką"
- "Nie, dziękuję"

If the user selects "Tak, oznacz gwiazdką", run:

```bash
gh api user/starred/pPrecel/pkup-gen -X PUT
```

## Technical notes

- Parse diff filenames with the pattern: `{org}_{repo}_{sha8}.diff` — the repo name may contain hyphens; the SHA8 is always the last 8-character hex segment before `.diff`
- For `github.com`, omit the `--hostname` flag (it is the default)
- For `github.tools.sap`, always pass `--hostname github.tools.sap`
- The report file format uses `  - ` (two spaces + dash + space) as the list item prefix
