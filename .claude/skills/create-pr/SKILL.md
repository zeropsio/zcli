---
name: create-pr
description: Prepare a GitHub PR for the current branch using .github/PULL_REQUEST_TEMPLATE.md, with fitting labels. Uses `gh` when available; otherwise outputs the filled PR description and suggested labels for the user to apply manually. Use when the user asks to "create a PR", "open a pull request", "make a PR", etc.
model: sonnet
---

# create-pr

Prepare a PR against `main` for the current branch, filling in the project's PR template and picking fitting labels.

Check whether `gh` is available (`command -v gh`). If yes, create the PR directly. If not, output the filled description and suggested labels for the user to apply manually on GitHub. Do not pressure the user to install `gh`.

## Steps

1. **Gather context** (parallel where possible):
   - `git status` (no `-uall`)
   - `git diff <base>...HEAD` to see the full set of changes on this branch (see step 2 for `<base>`)
   - `git log <base>..HEAD --oneline` to see commits and their conventional-commit prefixes
   - Check the branch is pushed and tracks a remote; if not, push with `-u` (only if `gh` will be used to create the PR)

2. **Pick the base branch.** Default to `main`. Before continuing, look for signals the PR might actually target a feature branch:
   - The current branch's upstream (`git rev-parse --abbrev-ref @{upstream}`) points to a non-main remote branch
   - `git merge-base --is-ancestor main HEAD` fails (HEAD is not based on the current tip of main but on some other branch)
   - The branch name suggests stacking (e.g. `feat/foo-part2` while `feat/foo-part1` exists)

   If any of those is true, ask the user which base to target before proceeding. Otherwise confirm `main` in one short sentence and continue without a prompt. Use the chosen base everywhere `<base>` appears below and pass it to `gh pr create --base <base>`.

3. **Verify before creating.** Check whether the branch touches Go code with `git diff --name-only <base>...HEAD | grep -E '\.go$|^go\.(mod|sum)$'`.

   - If Go files changed, run `make test` (full Go test suite including in-process integration tests under `cmd/...`) then `make lint`. If either fails, stop, show the user the failure, and do not proceed to PR creation.
   - If no Go files changed (docs, markdown, CI config, skills, etc.), skip both. State that explicitly in one short line so the user knows you skipped intentionally.

   Use the actual results (or the explicit skip) in the Test plan section below - do not paraphrase or invent.

4. **Determine Type** from the conventional-commit prefixes on the branch's commits:
   - `feat` → feat (user-visible new behavior)
   - `fix` → fix (bug fix)
   - `refactor` → refactor (no behavior change)
   - `docs` → docs (documentation only)
   - `chore` → chore (tooling, deps, housekeeping)
   - `ci` → ci (CI/CD pipeline)
   - `test` → test (tests only)

   If a branch mixes types (e.g. one `feat` plus a `test`), pick the dominant one - usually the highest in this priority order: feat > fix > refactor > docs > chore > ci > test.

5. **Draft the PR title.** Format rules:
   - **Sentence case, descriptive noun phrase.** Capitalize the first word, leave the rest lowercase unless it's a proper noun or technical identifier (e.g. `zcli`, `--check`). The title describes *what the PR is*, not what it does. Examples: `PR template revamp and create-pr skill`, `Race fix in version cache writer`, `--check flag for zcli upgrade`.
   - **Under 70 characters.** Avoids truncation in GitHub list views.
   - **No conventional-commit prefix.** Those belong on commits, not the title.
   - **No trailing period.** Title is a phrase, not a sentence.
   - **No backticks or markdown.** GitHub does not render markdown in titles; they appear literally.
   - **No issue number in the title.** Use `Closes #N` in the Description instead.

6. **Fill the PR body** from `.github/PULL_REQUEST_TEMPLATE.md`. Uncomment the chosen Type line(s), write Description, Test plan, Breaking changes. Drop the example block at the bottom. If the branch closes an issue, add `Closes #N` inside the Description (no need for a separate Closes section).

7. **Map Type to labels** for `gh pr create --label`:
   - feat → `enhancement`
   - fix → `bug`
   - refactor → `refactor`
   - docs → `documentation`
   - chore → `chore`
   - ci → `ci`
   - test → `test`
   - Additionally add `dependencies` if commits touch `go.mod`/`go.sum`/`package.json`/`flake.nix`.

8. **Create or output the PR**:

   - **With `gh` available:** create the PR with `gh pr create`, passing labels via `--label`, body via HEREDOC. Example shape:

     ```bash
     gh pr create \
       --title "Short human summary" \
       --label feat-mapped-label \
       --body "$(cat <<'EOF'
     ## Type

     **feat** - user-visible new behavior

     ## Description

     ...

     ## Test plan

     ...

     ## Breaking changes

     None.
     EOF
     )"
     ```

     Return the PR URL.

   - **Without `gh`:** output the filled PR description in a fenced code block (so it can be copy-pasted into the GitHub "New PR" form), the suggested title on its own line, and the suggested labels as a comma-separated list. Don't try to push the branch on the user's behalf in this path; let the user handle the GitHub side.

## Hard rules

- NEVER append a `Co-Authored-By: Claude ...` trailer to anything (PR body, commits).
- NEVER use `--no-verify`, `--no-gpg-sign`, or skip hooks.
- NEVER force-push without explicit user request.
- If `gh pr create` fails because labels don't exist on the repo, list current labels with `gh label list` and ask the user before creating new ones.
- Don't fill in fake test results. If you didn't run the tests in this session, write what the submitter still needs to run, or ask the user what they ran.