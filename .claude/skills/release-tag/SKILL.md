---
name: release-tag
description: Creates a new release for Nidus Dashboard — updates all documentation, then triggers the GitHub Actions Release workflow. Lists existing tags, suggests the next version, and asks for user confirmation.
user_invocable: true
argument: Optional tag version (e.g., "/release-tag v0.1.0")
---

You create releases for Nidus Dashboard by updating documentation and triggering the GitHub Actions `Release` workflow.

**The workflow handles everything:** tag creation, Docker image build, GitHub release, and Gotify notification. You must NEVER create tags or releases manually.

## Step 1: Gather Context (parallel)

Run all in parallel:

- `git tag --sort=-v:refname | head -10` — list recent tags
- `git log --oneline $(git tag --sort=-v:refname | head -1)..HEAD --no-merges` — commits since last tag (or all if no tags)
- `gh run list --workflow=Release --repo tdebuilt/Nidus-Dashboard --limit 3` — recent workflow runs

## Step 2: Determine Next Version

### If the user provided an argument:

Use it as the proposed version. Ensure it starts with `v` (e.g., `v0.2.0`).

### If no argument was provided:

Suggest the next version based on existing tags:
- Latest tag `v0.1.0` → suggest `v0.1.1` (patch), `v0.2.0` (minor), `v1.0.0` (major)
- No existing tags → suggest `v0.1.0`

### Pre-release versions

Tags containing a hyphen (e.g., `v0.3.0-beta.1`, `v0.3.0-rc.1`) are treated as pre-releases:
- The Docker image will NOT be tagged as `:latest` (only the versioned tag)
- The GitHub release should be marked as a prerelease (`gh release edit <tag> --prerelease`)
- Users on `:latest` are not affected

Examples: `v0.3.0-alpha.1`, `v0.3.0-beta.1`, `v0.3.0-rc.1`

## Step 3: Ask for Confirmation

Use `AskUserQuestion` to confirm. Show:
- The proposed version
- The latest existing tag (if any)
- Number of commits since last tag
- Summary of changes

## Step 4: Update Documentation

Review and update all documentation files to match the current state of the codebase.

### 4a. CHANGELOG.md

1. Read `CHANGELOG.md`
2. Move the `[Unreleased]` content into a new `[X.Y.Z] - YYYY-MM-DD` section
3. Reset `[Unreleased]` to empty

### 4b. ROADMAP.md

1. Read `ROADMAP.md`
2. Check each item in "Planned" against the codebase
3. If an item is implemented (code exists, tests pass), move it to "Completed" with `[x]`
4. Do NOT add new items to "Planned" — that's a manual decision

### 4c. SPEC.md

1. Compare the widget list in SPEC.md against `web/src/lib/widgetRegistry.ts`
2. Check that all features mentioned in the codebase are documented
3. Update any outdated sections (don't rewrite from scratch — make targeted edits)

### 4d. README.md

1. Verify badges point to the correct repo (`tdebuilt/Nidus-Dashboard`)
2. Check feature list matches current widgets
3. Verify installation instructions (Docker image, binary, desktop)

### 4e. CONTRIBUTING.md

1. Verify project structure matches actual directory layout
2. Check commands (`make lint`, `make test`, `make setup`, etc.) are current
3. Verify Node/Go version requirements match `package.json` and `go.mod`

### 4f. .todo/TESTING.md

1. Compare test cases against implemented features
2. Add missing test cases for new features, remove ones for removed features
3. Keep the same format: `## N. Section Name` with `- [ ]` checkboxes

### 4g. i18n Template

Check if `web/src/lib/i18n/fr.json` and `docs/i18n-template.json` have the same keys. If not, regenerate the template:

```python
python3 -c "
import json
with open('web/src/lib/i18n/fr.json') as f:
    fr = json.load(f)
def empty(d):
    return {k: empty(v) if isinstance(v, dict) else '' for k, v in d.items()}
with open('docs/i18n-template.json', 'w') as f:
    json.dump(empty(fr), f, indent=2, ensure_ascii=False)
    f.write('\n')
"
```

### 4h. Commit Documentation Updates

Commit and push all documentation changes before triggering the release.

## Step 5: Trigger the Release Workflow

```bash
gh workflow run Release --repo tdebuilt/Nidus-Dashboard -f version=<tag>
```

This single command triggers the GitHub Actions workflow which will:
1. Run all tests (Go + frontend + E2E)
2. Create the git tag
3. Build and push Docker image (`ghcr.io/tdebuilt/nidus-dashboard:<tag>` + `:latest`)
4. Create the GitHub release with auto-generated notes
5. Send Gotify notification

## Step 6: Confirm Workflow Started

```bash
gh run list --workflow=Release --repo tdebuilt/Nidus-Dashboard --limit 1
```

Display the run URL so the user can follow progress.

## Step 7: Wait for Workflow Completion and Update Release Notes

The workflow creates a release with auto-generated notes. After it completes, update the release description with a concise summary (~20 lines max).

1. Wait for the workflow to finish (poll with `gh run view <run-id> --repo tdebuilt/Nidus-Dashboard`)
2. Once complete, read the CHANGELOG.md section for this version
3. Update the release notes using `gh release edit` following this format:

```
## <tag> — <short title> (e.g., "Initial Release", "Grafana widget & service improvements")

<1-2 sentence summary of what this release brings.>

### Docker

\```
docker pull ghcr.io/tdebuilt/nidus-dashboard:<tag>
\```

### Highlights

- **Key feature 1** — brief description
- **Key feature 2** — brief description
- (8-10 bullet points max, condensed from CHANGELOG.md)

Full changelog: [CHANGELOG.md](https://github.com/tdebuilt/Nidus-Dashboard/blob/main/CHANGELOG.md)
```

Rules for release notes:
- **~20 lines max** — concise, not exhaustive
- Condense CHANGELOG items into grouped highlights (e.g., list all widgets on one line)
- Always include the `docker pull` command with the versioned tag
- Always link to CHANGELOG.md for full details

## Step 8: Summary

Display:

```
Release <tag> created.
Workflow: <run-url>
Docker image: ghcr.io/tdebuilt/nidus-dashboard:<tag>
Release: https://github.com/tdebuilt/Nidus-Dashboard/releases/tag/<tag>
```

## Important Rules

- NEVER create tags manually (`git tag`) — the workflow does it
- NEVER create GitHub releases manually (`gh release create`) — the workflow does it
- ALWAYS ask for user confirmation before triggering
- ALWAYS update documentation before triggering
- ALWAYS update release notes after workflow completes with Docker image link and changes summary
- Only update doc files that actually need changes — don't rewrite for style
- All documentation must be in English
- Never mention Claude, AI, or automation in any documentation
- If the workflow trigger fails, display the error and STOP
