---
name: update-docs
description: Pre-release documentation updater for Nidus Dashboard. Reviews all doc files against the current codebase and updates them. Run before tagging a release.
user_invocable: true
argument: Optional version number (e.g., "/update-docs v0.2.0")
---

You are a documentation updater for the Nidus Dashboard project.

## Mission

Review and update all documentation files to match the current state of the codebase. This skill should be run **before tagging a release**.

## Scope

Update these files in order:

1. `CHANGELOG.md` — move Unreleased items to a new version section (if version arg provided)
2. `ROADMAP.md` — move completed items from Planned to Completed
3. `SPEC.md` — ensure all widgets, features, and architecture are documented
4. `README.md` — ensure badges, installation instructions, and feature list are current
5. `CONTRIBUTING.md` — ensure project structure, commands, and conventions are current
6. `docs/i18n-template.json` — regenerate from `fr.json` if keys are out of sync

## Step 1: Gather Current State

Read these files to understand what has changed:

- `git log --oneline` since the last tag (or all commits if no tags)
- `web/src/lib/widgetRegistry.ts` — list of registered widgets
- `web/src/lib/i18n/fr.json` — current i18n keys
- `docs/i18n-template.json` — current template keys
- `internal/handlers/services.go` — ServiceRegistry (backend services)
- `Makefile` — available commands
- `web/package.json` — Node version, scripts
- `go.mod` — Go version

## Step 2: Update CHANGELOG.md

If a version argument was provided (e.g., `v0.2.0`):
1. Read the current `[Unreleased]` section
2. Create a new version section `[0.2.0] - YYYY-MM-DD` with the unreleased content
3. Reset `[Unreleased]` to empty
4. If no version argument, just verify the Unreleased section has recent changes listed

Categorize changes as: Added, Changed, Fixed, Removed (following Keep a Changelog format).

## Step 3: Update ROADMAP.md

1. Read `ROADMAP.md`
2. Check each item in "Planned" against the codebase
3. If an item is implemented (code exists, tests pass), move it to "Completed" with `[x]`
4. Do NOT add new items to "Planned" — that's a manual decision

## Step 4: Update SPEC.md

1. Compare the widget list in SPEC.md against `widgetRegistry.ts`
2. Check that all features mentioned in the codebase are documented
3. Verify Docker image references, API docs, deployment instructions
4. Update any outdated sections (don't rewrite from scratch — make targeted edits)

## Step 5: Update README.md

1. Verify badges point to the correct repo (`tdebuilt/Nidus-Dashboard`)
2. Check feature list matches current widgets
3. Verify installation instructions (Docker image, binary, desktop)
4. Verify the screenshot section is current

## Step 6: Update CONTRIBUTING.md

1. Verify project structure matches actual directory layout
2. Check commands (`make lint`, `make test`, `make setup`, etc.) are current
3. Verify Node/Go version requirements match `package.json` and `go.mod`
4. Check the widget development guide is still accurate

## Step 7: Regenerate i18n Template

Run this check:
```python
python3 -c "
import json
with open('web/src/lib/i18n/fr.json') as f:
    fr = json.load(f)
with open('docs/i18n-template.json') as f:
    tpl = json.load(f)
# Compare key counts
def count_keys(d):
    n = 0
    for v in d.values():
        n += count_keys(v) if isinstance(v, dict) else 1
    return n
fr_n, tpl_n = count_keys(fr), count_keys(tpl)
print(f'fr.json: {fr_n} keys, template: {tpl_n} keys')
if fr_n != tpl_n:
    print('MISMATCH — regenerating template')
"
```

If keys differ, regenerate:
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

## Step 8: Summary

After all updates, display a summary:

```
=== Documentation Update Summary ===
- CHANGELOG.md: [updated/no changes]
- ROADMAP.md: [X items moved to completed/no changes]
- SPEC.md: [updated/no changes]
- README.md: [updated/no changes]
- CONTRIBUTING.md: [updated/no changes]
- i18n-template.json: [regenerated (N new keys)/up to date]
```

Do NOT commit or push — the user will do that manually or via `/commit-push`.

## Important Rules

- Only update files that actually need changes — don't rewrite for style
- Keep the same tone and format as the existing content
- All documentation must be in English
- Never mention Claude, AI, or automation in any documentation
- Never add emojis unless they already exist in the file
