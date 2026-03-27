---
name: release-desktop
description: Builds desktop apps (Linux, macOS, Windows) for an existing Nidus release by triggering the GitHub Actions Release Desktop workflow.
user_invocable: true
argument: Optional tag version (e.g., "/release-desktop v0.1.0")
---

You build desktop apps for Nidus Dashboard by triggering the GitHub Actions `Release Desktop` workflow. The binaries are attached to an **existing** GitHub release.

## Step 1: List Existing Tags

```bash
git tag --sort=-v:refname | head -10
```

Only existing tags are valid targets — this skill does NOT create tags or releases.

## Step 2: Determine Target Tag

### If the user provided an argument:

Verify the tag exists locally and on remote. If it does not exist, display an error and STOP.

### If no argument was provided:

Show the list of existing tags and ask the user to pick one using `AskUserQuestion`.

## Step 3: Verify Release Exists

```bash
gh release view <tag> --repo tdebuilt/Nidus-Dashboard
```

If no release exists for this tag, display an error and STOP. The user must run `/release-tag` first.

## Step 4: Ask for Confirmation

Use `AskUserQuestion` to confirm. Show:
- The target tag
- The existing release title
- Platforms that will be built: Linux (.deb, .AppImage), macOS Intel + Apple Silicon (.dmg), Windows (.msi)

## Step 5: Trigger the Desktop Workflow

```bash
gh workflow run "Release Desktop" --repo tdebuilt/Nidus-Dashboard -f tag=<tag>
```

Note: the input field is `tag`, not `version`.

## Step 6: Confirm Workflow Started

```bash
gh run list --workflow="Release Desktop" --repo tdebuilt/Nidus-Dashboard --limit 1
```

Display the run URL so the user can follow progress.

## Step 7: Summary

Display:

```
Desktop build for <tag> triggered.
Workflow: <run-url>
Binaries will be attached to: https://github.com/tdebuilt/Nidus-Dashboard/releases/tag/<tag>
```

## Important Rules

- ONLY target tags that already exist — NEVER create tags
- ONLY target tags that have a GitHub release — STOP if no release found
- ALWAYS ask for user confirmation before triggering
- If the workflow trigger fails, display the error and STOP
