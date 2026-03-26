---
name: next-task
description: "Implement a task for Nidus Dashboard. Three modes: (1) no args → picks next pending task from roadmap, (2) task reference (e.g. 'plugin system') → finds and implements that roadmap task, (3) free-form description (e.g. 'add a reset button') → implements it as a custom task. Use when user says 'tâche suivante', 'next task', or describes something to implement."
user-invocable: true
disable-model-invocation: false
effort: high
argument-hint: "[optional: task keyword (plugin system), or free-form description]"
---

# Next Task Workflow

You are implementing a task for the Nidus Dashboard project. Follow this workflow strictly.

## Step 0 — Find the task

### Mode A — No arguments
1. Read `ROADMAP.md` and find the first uncompleted item in the "Planned" section
2. If all planned tasks are done, inform the user

### Mode B — Roadmap reference (`$ARGUMENTS` matches a task)
1. Read `ROADMAP.md`
2. Find the task matching the argument: section name or keyword (e.g. `plugin system`)
3. If found, implement that task

### Mode C — Custom task (`$ARGUMENTS` doesn't match any roadmap task)
1. Treat `$ARGUMENTS` as a free-form task description
2. This is a custom task — it won't have an entry in the roadmap
3. Skip Step 5 (no roadmap status to update) but still follow all other steps

**If the task requires external input** (screenshots, manual testing, external API keys, translations by native speakers), mark it as blocked with a reason, skip it, and move to the next task. (Only applies to Modes A and B.)

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

Share the plan with the user. **Do NOT wait for confirmation** — proceed immediately unless the task is architecturally significant (new service, DB schema change, breaking change).

## Step 3 — Implement

Execute each sub-task:
- Follow existing code conventions (check nearby files for patterns)
- Backend Go: packages in `internal/`, English code
- Frontend Svelte: components in `web/src/lib/components/`, use the widget registry for new widgets
- i18n: update all locale files under `web/src/lib/i18n/`
- Keep changes minimal and focused — don't refactor unrelated code

## Step 4 — Test

- Run frontend lint: `cd web && npm run lint`
- Run frontend tests: `cd web && npm test -- --run`
- Run Go lint: `make lint-go`
- If tests fail, fix the issue and re-run
- If you created new components, write tests for them

## Step 5 — Update task status (Modes A & B only)

1. In `ROADMAP.md`, move the completed item from "Planned" to "Completed" with `[x]`
2. If a task was blocked, note the reason

**Skip this step for Mode C (custom tasks).**

## Step 6 — Build & commit

- Rebuild the app: `docker compose up --build -d`
- Stage only relevant files (no `git add -A`)
- Write a descriptive commit message in English
- Push to origin

## Step 7 — Report

Output a summary of what was done:
- Which task was completed
- Files created/modified
- Test results
- Any issues encountered
- What the next pending task is (for Modes A & B)

## Important rules

- **One task per invocation** — complete one task fully, then stop
- **Never skip tests** — always run the test suite
- **Never break existing functionality** — if tests fail after your changes, fix them
- **Follow CLAUDE.md conventions** — English code, all documentation in English
- **Use the widget registry** when adding new widgets — don't hardcode
- **Commit message style**: lowercase prefix (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`)
