---
name: new-release
description: Cut a new zcli release end-to-end. Two phases - first opens a `release/vX.Y.Z` PR adding the `CHANGELOG.md` entry, then (once the user merges that PR) creates the GitHub Release on main which triggers `.github/workflows/release.yml`, watches goreleaser + npm publish + smoke tests, and reports the outcome. Use when the user says "cut a release", "new release", "release", "ship a new version", "tag and release", "release an rc", "promote to stable", "publish v1.0.x", or anything similar - even if they don't name a specific version. **Also re-trigger when the user says the release PR is merged**, "continue the release", "finish the release", "PR is in - ship it", or re-asks about a release they already started; in that case skip to step 4 (Resume after the PR merges).
model: sonnet
---

# new-release

Cut a new zcli release end-to-end.

## What "release" means in this repo

A release is **not** a `git tag` + `git push`. The release pipeline (`.github/workflows/release.yml`) triggers on the `release` event (`released` / `prereleased`), **not** on tag pushes. So:

- The action that actually starts a release is `gh release create vX.Y.Z`. That creates the GitHub Release object and emits the event the workflow listens for.
- Pushing a bare tag does nothing useful and leaves you with a dangling tag that pollutes the list.

The pipeline has three downstream jobs you care about:
1. **goreleaser** - cross-builds binaries, uploads release assets, generates checksums.
2. **publish-npm** - downloads `*-npm.tar.gz` assets, builds and publishes platform packages, then publishes the main `@zerops/zcli` package under dist-tag `latest` (stable) or `next` (pre-release).
3. **smoke-test** - installs the just-published npm package across linux/mac/windows and npm/bun, runs `zcli version`.

Stable vs pre-release matters because it picks the npm dist-tag, and because consumers of `@zerops/zcli` (without `@next`) only see stable. **Pre-release tags use a `-rc<N>` suffix** (e.g. `v1.0.68-rc0`); stable tags are plain `vMAJOR.MINOR.PATCH`.

## Pause before every mutating `gh` call

Read-only `gh` calls (`gh run list`, `gh run view`, `gh release view`, `gh release list`) can run freely - they don't change state. Anything that creates, deletes, or re-runs **must pause for explicit user confirmation first**, even if the user already said "cut a release" at the top of the conversation. Releases are public, npm versions are immutable, and CI minutes cost money - the cost of an accidental `gh release create` is much higher than the cost of one extra "y/n" round-trip.

The mutating calls in this skill are:
- `git push` of the `release/vX.Y.Z` branch and `gh pr create` for the release PR (step 3, **stable only**)
- `gh release create ...`
- `gh release delete ...`
- `gh run rerun ...`

Note: for stable releases the CHANGELOG commit does **not** land on `main` directly - it goes through a release PR that a human merges, and that merge is the gate. Pre-releases skip the PR entirely (see step 3) and go straight from version selection to `gh release create --prerelease`; the rc isn't worth a separate changelog entry, and the stable release that eventually promotes it carries one entry covering the whole rc series.

For each, show the user the **exact command you're about to run** (full flags, no abbreviation) on its own line, state in one sentence what it will do and what's irreversible about it, and wait for an explicit yes. "Looks good?" / "ready?" is not enough - quote the command. If the user's initial request was specific enough to pin down the command (e.g. "promote v1.0.68-rc0 to stable" → unambiguous `gh release create v1.0.68 ...`), you still pause - the user is confirming the exact command, not the intent.

## Steps

### 1. Pre-flight

Run in parallel:
- `git rev-parse --abbrev-ref HEAD` - confirm on `main` (refuse otherwise unless the user explicitly says they want to release off another branch; ask).
- `git status --porcelain` - working tree must be clean.
- `git fetch --tags origin` then `git rev-list --left-right --count HEAD...origin/main` - local must not be behind `origin/main` (if behind, pull; if ahead, ask).
- `gh run list --workflow=main.yml --branch=main --limit 1 --json status,conclusion,headSha,displayTitle` - the latest `main` workflow must be `completed` + `success`, and its `headSha` must match `git rev-parse HEAD`. If the latest run is still running or failed, stop and tell the user; do not release on top of red or unknown state.

If any check fails, stop and surface the exact problem. Do not "fix" it silently (don't auto-pull, don't auto-stash).

### 2. Decide the version

Pull the last few tags and recent commits, then propose:

```bash
git tag --sort=-creatordate | head -10
# also get the latest stable specifically:
git tag --sort=-creatordate | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1
git log <last-stable>..HEAD --oneline
```

Then propose **the next version and whether it's a pre-release**, with one-sentence reasoning, and ask the user to confirm or override. Default heuristics:

- If the last tag is a pre-release (`-rcN`) and nothing has changed since, suggest **promoting** to the same base version as a stable release (e.g. `v1.0.68-rc0` → `v1.0.68`).
- If the last tag is a pre-release and there are new commits since, suggest the **next rc** (`-rc{N+1}`).
- If the last tag is stable and there are new commits, suggest a **new rc0** at the next patch (`v1.0.68` → `v1.0.69-rc0`), unless the commits are clearly trivial enough that the user wants to go straight to stable. Bumping minor/major requires explicit user direction - don't infer it from commit messages.
- If you can't tell, ask. Never guess silently.

Always show the user: the proposed version, whether it's pre-release, the last stable, the last pre-release if different, and a short list of the commits that would be in the release.

### 3. Open a release PR with the CHANGELOG entry (stable only)

**Pre-releases skip this step entirely.** If the release you're cutting is an rc (or any `-suffix`ed tag), jump directly to step 5 - rcs are throwaway iterations, the project has never tracked them in `CHANGELOG.md`, and forcing a PR per rc would make the rc → rc → stable cycle painful. The eventual stable release will carry one CHANGELOG entry covering everything from the previous stable through all the rcs in between.

For a stable release, the new entry **does not go to `main` directly**. It lands via a release PR: branch → push → PR → human merges → you resume. The merge is the gate the human uses to approve the release; do not bypass it.

**Draft the entry from the commits since the last *stable* release** (not the last tag - the last tag is usually the most recent rc, which would understate the diff). Use `git log <last-stable>..HEAD --pretty='%h %s'` and classify by conventional-commit prefix:

| Prefix                                               | Section                 |
| ---------------------------------------------------- | ----------------------- |
| `feat:` / `feat(...)`                                | `### Added`             |
| `fix:` / `fix(...)`                                  | `### Fixed`             |
| `refactor:` / `perf:` / breaking                     | `### Changed`           |
| `chore:`, `ci:`, `test:`, `docs:`, `lint:`, `style:` | skip (not user-visible) |

Rewrite each line as a short user-facing sentence - drop the prefix and scope, expand jargon, present-tense imperative. The source commit subject is for you to understand intent, not to copy verbatim. If a commit message is too terse to summarize confidently, read the diff or the PR title before guessing. Group adjacent commits that describe one feature into a single bullet. Drop bullets that aren't meaningful to a user installing the new version (purely internal refactors that change no behavior, for instance).

Keep the existing file's format exactly:

```markdown
## [vX.Y.Z] - YYYY-MM-DD

### Added
- short user-facing description

### Fixed
- ...
```

Date is today, in `YYYY-MM-DD`. Insert immediately after the file's intro paragraph (above the most recent existing entry). If only one section has entries (e.g. only fixes), only emit that section - don't include empty `### Added` blocks.

**Show the drafted entry to the user and let them edit before you write anything.** Wait for explicit approval.

After approval, open the release PR. Show the user the exact commands first (this is a mutating sequence - see the pause section above) and wait for confirmation, then run:

```bash
git checkout -b release/vX.Y.Z          # branch off the current main
# write the CHANGELOG.md change here
git add CHANGELOG.md
git commit -m "docs(changelog): vX.Y.Z"
git push -u origin release/vX.Y.Z
gh pr create \
    --base main \
    --head release/vX.Y.Z \
    --title "release: vX.Y.Z" \
    --body "$(cat <<'EOF'
Changelog entry for vX.Y.Z. Merging this PR is the approval gate for the release - once it merges, the release skill resumes and cuts the GitHub Release on the new `main` HEAD.

<paste the same CHANGELOG entry here as the PR body so reviewers see it inline>
EOF
)"
```

**Then stop and tell the user explicitly: "PR opened at <url>. Merge it when you're ready, then come back and tell me to continue the release."** Do not attempt to merge the PR yourself - even if you have permission, the merge is the human's deliberate sign-off on what's about to ship.

### 4. Resume after the PR merges

The user comes back saying something like "the release PR is merged", "continue the release", "PR is in - finish the release", or just re-invokes the skill with the same version. Pick up here:

1. **Find and verify the release PR is merged.** `gh pr list --search "release: vX.Y.Z in:title" --state all --json number,state,mergeCommit,baseRefName` - the PR must be `MERGED` and target `main`. If still open, stop and tell the user.
2. **Sync local main and re-run pre-flight.** `git checkout main && git pull --ff-only origin main`. Then redo the step 1 checks against the new HEAD - clean tree, not behind origin, and **`main` CI must be green at the merge SHA.** `gh run list --workflow=main.yml --branch=main --limit 1 --json status,conclusion,headSha` and assert `headSha` matches `git rev-parse HEAD`. If CI is still running, wait or tell the user to come back; if it failed, stop.
3. **Check for unexpected drift.** Run `git log <merge-sha>..HEAD --oneline` - if anything besides the merge commit landed on main between PR creation and now, those commits will also be in the release. Surface them to the user before proceeding so they can decide whether the changelog needs another PR to cover the additions, or whether it's fine to ship them under the same version.
4. Delete the local `release/vX.Y.Z` branch (`git branch -d release/vX.Y.Z`) - housekeeping, not safety-critical.

Only after these checks pass, proceed to step 5.

### 5. Create the release

#### 5a. Draft the release notes

Do not use `--generate-notes`. Write the notes yourself - they're the first thing a user sees on the release page, and the auto-generated PR dump is noisy and doesn't categorize.

**Fetch the PR list for the release range** using the API endpoint that backs `--generate-notes`:

```bash
gh api repos/zeropsio/zcli/releases/generate-notes \
    -F tag_name=vX.Y.Z \
    -F target_commitish=main \
    -F previous_tag_name=<previous-tag> \
    --jq .body
```

`<previous-tag>` is:
- For a **stable** release: the **previous stable** tag (so promoting `v1.0.68-rc2` → `v1.0.68` covers everything since `v1.0.67`, including all the rc-only commits the user never saw).
- For a **pre-release**: the previous tag of any kind (most recent rc, or the previous stable if this is rc0). The point is to show only what's new since the last published thing.

The returned body has one bullet per merged PR with title, author, and URL. Use it as the **source list** only - don't paste it as-is. Filter out the release PR itself if it's in the list (its title starts with `release: v`), since "Add changelog entry" isn't a release-worthy bullet.

**Format the notes like this:**

```markdown
<one or two sentence summary of what this release is about - the headline, in your own words. For an rc, say "Pre-release for testing X before stable." For a stable that promotes an rc series, name the theme.>

### Added
- Short user-facing summary of the change ([#NUM](https://github.com/zeropsio/zcli/pull/NUM))

### Fixed
- ...

### Changed
- ...

**Full Changelog**: https://github.com/zeropsio/zcli/compare/<previous-tag>...vX.Y.Z
```

Classify each PR by its dominant change type (use the PR title's conventional-commit prefix as the signal: `feat:` → Added, `fix:` → Fixed, `refactor:`/`perf:`/breaking → Changed; skip `chore:`/`ci:`/`test:`/`docs:`/`lint:`/`style:` unless a "chore" PR is actually the headline of the release, in which case it deserves a bullet under whichever section fits). Each bullet should be a **rewritten** short sentence, not the raw PR title - drop the prefix, expand jargon, make it about what the user gets. Reference the PR with `[#NUM](url)` so readers can click through. Group adjacent related PRs into one bullet with multiple refs (`([#252](...), [#254](...))`).

For a **stable promoting an rc series**: the notes should describe the *cumulative* changes since the previous stable, not just what's different from the last rc. Most readers skipped the rcs.

**Show the drafted notes to the user and wait for explicit approval before continuing.** Save the approved version to a file (e.g. `/tmp/release-notes-vX.Y.Z.md`) so the next command can read it back without escaping headaches.

#### 5b. Create the release

Build the exact `gh release create` command and show it to the user verbatim before running it (see the pause-before-mutating-calls section above). Templates:

```bash
# Stable:
gh release create vX.Y.Z --target main --title vX.Y.Z --notes-file /tmp/release-notes-vX.Y.Z.md

# Pre-release:
gh release create vX.Y.Z-rcN --target main --title vX.Y.Z-rcN --prerelease --notes-file /tmp/release-notes-vX.Y.Z-rcN.md
```

Notes:
- `--notes-file` takes the path to the notes you drafted in 5a. Don't use `--notes "..."` with inline shell-escaped markdown - backticks and parens in PR titles will bite you.
- The GitHub release notes (your hand-written 5a output) and `CHANGELOG.md` are deliberately separate: the release page is dense and PR-linked for someone evaluating the upgrade, the changelog is plain prose for the historical record. Same source PRs, different framing.
- `--target main` pins the release to the current `main` SHA (which now includes the merged release PR from step 3). Don't omit `--target` - leaving it unset lets the API pick a default that may surprise you.
- The `v` prefix on the git tag is load-bearing: the npm publish step strips it (`VERSION#v`) to produce a plain semver for `package.json`. Stable tags must match `^v\d+\.\d+\.\d+$` exactly; pre-release tags must use `-rc\d+`. Don't invent other suffixes (`-beta`, `-alpha`) without checking with the user, because the workflow's `prereleased` dispatch logic only cares about the GitHub release flag, not the suffix - but the dist-tag heuristic in `release.yml` does (`'next' || 'latest'`), so other suffixes would still go out as `next` which is fine, but is a deviation from project convention worth confirming.

### 6. Watch the pipeline

Immediately after `gh release create` returns, find the triggered workflow run and follow it:

```bash
# Find the run for this release tag. It may take a few seconds to appear.
gh run list --workflow=release.yml --event=release --limit 5 \
    --json databaseId,displayTitle,status,conclusion,headBranch,createdAt
```

Pick the run whose `displayTitle` (or `headBranch`) matches the tag you just created. Then `gh run watch <id> --exit-status` to block until it finishes, or poll with `gh run view <id> --json jobs` if you want a structured per-job status.

Report each job's outcome to the user as it completes (or at least once at the end):
- **goreleaser** failed → release exists but assets are missing. Do not delete the release silently; tell the user and ask before re-running.
- **publish-npm** failed → goreleaser succeeded so assets are there, but npm is incomplete. Re-running the job is safe: platform publishes skip versions already on the registry, and the main package only publishes after all platform packages exist (this is intentional - see the comment in `release.yml`).
- **smoke-test** failed on a single matrix cell → the release went out but a specific platform/manager combo is broken. Surface which one.

### 7. Verify

After the pipeline ends successfully:

```bash
V="${TAG#v}"
gh release view "$TAG" --json assets --jq '.assets[].name'   # confirm goreleaser uploaded
npm view "@zerops/zcli@$V" version dist-tags                  # confirm the right dist-tag
```

For a stable release: expect `dist-tags.latest == $V`.
For a pre-release: expect `dist-tags.next == $V` and `dist-tags.latest` unchanged.

Report the release URL (`gh release view $TAG --json url --jq .url`) at the end so the user has a link.

## If something goes wrong

- **You created the release with the wrong version/notes/prerelease flag** and nothing has published yet: `gh release delete <tag> --cleanup-tag --yes` removes both. Show the user the exact command and wait for confirmation - delete is irreversible (the tag goes with it), and if `publish-npm` has *already* pushed to the registry you must not delete, because npm versions are immutable and you'd end up with a tag-less version on npm that the workflow can never rebuild. Re-check the run state right before deleting.
- **`publish-npm` partially published** (some platform packages went out, main didn't): re-run the `publish-npm` job. Safe because of the idempotent skip-if-present check, but still a mutating action - show the user the exact `gh run rerun <id> --failed` command and wait for confirmation before running it.
- **goreleaser failed**: usually a build error. Fix on `main`, delete the release+tag, re-cut.
- **The user wants to abort mid-release** (after `gh release create` but before publish): if goreleaser hasn't started yet you can delete the release+tag; once npm has anything published you cannot undo it - your only option is to ship a new version forward.

## What not to do

- Don't `git tag vX.Y.Z && git push --tags` - that creates a tag without a release and does not trigger the pipeline.
- Don't skip the `CHANGELOG.md` update (step 3) even for a tiny release - the file is the user-facing record of what shipped, and gaps are noticed.
- Don't bump versions in `package.json` files manually - `tools/npm/scripts/build-platform-packages.mjs` and the `replace-in-files` step in `release.yml` do it from the tag.
- Don't run `goreleaser release` locally. Use `make goreleaser-snapshot` if you want to dry-run, but the real release happens in CI with `GITHUB_TOKEN` and the project's secrets.
- Don't skip the pre-flight green-CI check, even "just this once". The release pipeline doesn't re-run tests; it builds from the tagged SHA assuming it's known-good.