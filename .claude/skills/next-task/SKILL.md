---
name: next-task
description: "Implement a task for Nidus Dashboard. Three modes: (1) no args → picks next pending task from roadmap, (2) task reference (e.g. '4.7', 'widget RSS') → finds and implements that roadmap task, (3) free-form description (e.g. 'ajouter un bouton de reset') → implements it as a custom task. Use when user says 'tâche suivante', 'next task', or describes something to implement."
user-invocable: true
disable-model-invocation: false
effort: high
argument-hint: "[optional: task number (4.7), keyword (widget RSS), or free-form description]"
---

# Next Task Workflow

You are implementing a task for the Nidus Dashboard project. Follow this workflow strictly.

## Step 0 — Find the task

### Mode A — No arguments
1. Read `planning/OPEN_SOURCE_TASKS.md` first (Phase 1 has priority)
2. If all Phase 1 tasks are done, read `planning/ROADMAP_TASKS.md` and find the first `[ ]` (pending) task

### Mode B — Roadmap reference (`$ARGUMENTS` matches a task)
1. Read `planning/ROADMAP_TASKS.md` (and `planning/OPEN_SOURCE_TASKS.md` if needed)
2. Find the task matching the argument: phase number (e.g. `4.7`), section name, or keyword (e.g. `widget RSS`)
3. If found, implement that task

### Mode C — Custom task (`$ARGUMENTS` doesn't match any roadmap task)
1. Treat `$ARGUMENTS` as a free-form task description
2. This is a custom task — it won't have a checkbox in the roadmap files
3. Skip Step 5 (no roadmap status to update) but still follow all other steps

**If the task requires external input** (screenshots, manual testing, external API keys, translations by native speakers), mark it as `[⏳ blocked]` with a reason, skip it, and move to the next `[ ]` task. (Only applies to Modes A and B.)

Announce to the user which task you're working on.

## Step 1 — Explore

Explore the codebase to understand:
- What files are relevant to this task
- What patterns/conventions are already in place
- What dependencies or related code exists
- What tests exist that might need updating

Use Grep, Glob, and Read to investigate. Be thorough — understanding the existing code prevents rework.

## Step 2 — Plan

Create a concrete list of sub-tasks. For each sub-task, specify:
- The file(s) to create or modify
- What exactly to change
- Any new dependencies needed

Share the plan with the user via Telegram (chat_id from the most recent Telegram message) or as text output. **Do NOT wait for confirmation** — proceed immediately unless the task is architecturally significant (new service, DB schema change, breaking change).

## Step 3 — Implement

Execute each sub-task:
- Follow existing code conventions (check nearby files for patterns)
- Backend Go: packages in `internal/`, English code, French UI strings
- Frontend Svelte: components in `web/src/lib/components/`, use the widget registry for new widgets
- i18n: update both `fr.json` and `en.json` (and all other locale files)
- Keep changes minimal and focused — don't refactor unrelated code

## Step 4 — Test

- Run frontend tests: `cd /home/claude-user/Nidus-Dashboard/web && npx vitest run`
- If tests fail, fix the issue and re-run
- If you created new components, write tests for them
- Backend Go tests can only run via Docker (not available locally) — ensure the code compiles by checking syntax

## Step 5 — Update task status (Modes A & B only)

1. In the task file (`planning/OPEN_SOURCE_TASKS.md` or `planning/ROADMAP_TASKS.md`), change `[ ]` to `[x]` for the completed task
2. Update the summary table counts (Finish +1, Pending -1)
3. If a task was blocked, change `[ ]` to `[⏳ blocked: reason]`

**Skip this step for Mode C (custom tasks).**

## Step 6 — Commit & push

- Stage only relevant files (no `git add -A`)
- Write a descriptive commit message in English
- Push to origin
- **Do NOT create a release** unless explicitly asked

## Step 7 — Report

Send a summary of what was done:
- Which task was completed
- Files created/modified
- Test results
- Any issues encountered
- What the next pending task is (for Modes A & B)

If Telegram is available, send the report there. Otherwise output as text.

## Important rules

- **One task per invocation** — complete one task fully, then stop
- **Never skip tests** — always run the test suite
- **Never break existing functionality** — if tests fail after your changes, fix them
- **Follow CLAUDE.md conventions** — no "Claude" in commits, no co-author, English code, French UI
- **Use the widget registry** when adding new widgets — don't hardcode
- **Commit message style**: lowercase prefix (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`)
