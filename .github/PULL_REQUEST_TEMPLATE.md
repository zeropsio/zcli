<!-- Thanks for the PR! -->

## Type

<!-- Uncomment the line(s) that apply. Match the conventional-commit prefix on your commits. -->

<!-- **feat** - user-visible new behavior -->
<!-- **fix** - bug fix -->
<!-- **refactor** - no behavior change -->
<!-- **docs** - documentation only -->
<!-- **chore** - tooling, deps, housekeeping -->
<!-- **ci** - CI/CD pipeline -->
<!-- **test** - tests only -->

## Description

<!-- What does this change, and why? Link any relevant context. -->

## Test plan

<!-- How did you verify this? Commands run, platforms tested (macOS / Linux / Windows), notable edge cases. Paste output if useful. -->

## Breaking changes

<!-- Removed/renamed flags, changed command output, changed exit codes, config format changes. Leave "None" if none. -->

None.

## Closes

<!-- Optional. e.g. Closes #123 -->

<!--
Don't forget to apply labels via the GitHub sidebar. Available:
bug, enhancement, refactor, documentation, chore, ci, test, dependencies,
duplicate, good first issue, help wanted
-->

<!--
=== Example (delete before submitting) ===

## Type

**feat** - user-visible new behavior

## Description

Adds `zcli upgrade --check` so users (and package managers) can detect whether
a newer release is available without performing the upgrade itself. Exits 0 if
up to date, non-zero with a structured message if an update exists.

## Test plan

- `go test ./src/cmd/... ./src/version/...` on macOS arm64
- Manual: `zcli upgrade --check` against a stubbed API URL returning a higher
  version, confirmed exit code 10 and the channel-aware hint
- Manual: same with the API URL returning the current version, confirmed exit 0

## Breaking changes

None.

## Closes

Closes #142
-->
