//go:build devel

// Package cmd's integration tests (this file documents them; the tests
// themselves live in *_integration_test.go and pushDeploy_*_test.go).
//
// All integration tests build under the `devel` build tag (so the
// production-version-check HTTP call from src/version is stubbed to a no-op)
// and run via `make test-integration`. They drive the CLI in-process through
// cmdBuilder.RunRootCmd, point the REST client at an httptest.Server, and
// isolate per-test state by setting ZEROPS_CLI_DATA_FILE_PATH,
// ZEROPS_CLI_LOG_FILE_PATH, and ZEROPS_CLI_YAML_FILE_PATH to per-test temp
// files. yamlReader's package-level cache is reset before and after each
// test. See integration_harness_test.go and pushDeploy_helpers_test.go for
// the shared scaffolding.
//
// # Tests by command
//
// ## zcli version (version_integration_test.go)
//
//   - TestVersionCommand
//     `zcli version` exits 0 and stdout starts with "zcli version ".
//
// ## zcli login (login_integration_test.go)
//
//   - TestLoginCommand_PersistsTokenAndRegion
//     `zcli login <token> --region-url ...` resolves the region from the
//     mocked region catalog, fetches GetUserInfo, and persists token + region
//     to cliStorage. The success message names the user.
//
// ## zcli project list (projectList_integration_test.go)
//
//   - TestProjectListCommand
//     `zcli project list` walks GetUserInfo → org filtering (ACTIVE only) →
//     PostProjectSearch, then renders a table that contains the project id,
//     name, and org name.
//
// ## zcli service deploy (serviceDeploy_integration_test.go)
//
//   - TestServiceDeployCommand_HappyPath
//     The deploy variant of push: same scope/yaml/validation/app-version flow
//     but the final request goes to /app-version/{id}/deploy (not
//     /build-and-deploy). Locks the existence of the variant and confirms
//     setup forwarding.
//
// ## zcli service push (servicePush_integration_test.go)
//
// ### Happy paths
//
//   - TestServicePushCommand_SetupAutoMatchesServiceName
//     Service name matches a setup in zerops.yaml → no --setup flag needed,
//     auto-match wins. Asserts the resolved setup is forwarded in the deploy
//     body.
//   - TestServicePushCommand_SetupSelectedByFlag
//     Service name doesn't match any setup → --setup explicitly selects one
//     from the yaml.
//   - TestServicePushCommand_VersionNameForwarded
//     --version-name v1.2.3 lands in the POST /app-version request body.
//
// ### Error paths / variant coverage
//
//   - TestServicePushCommand_MissingZeropsYaml
//     No zerops.yaml/.yml in working-dir → non-zero exit, error mentions the
//     yaml. No API call is made.
//   - TestServicePushCommand_EmptyZeropsYaml
//     Zero-byte zerops.yaml is treated as a hard error.
//   - TestServicePushCommand_InvalidWorkspaceState
//     --workspace-state=garbage fails client-side validation before any
//     archive work.
//   - TestServicePushCommand_NoSetupMatchNoFlagFailsInNonTTY
//     Non-TTY run with no auto-match and no --setup must fail with the
//     "please select with --setup" message — never silently invoke an
//     interactive selector that can't run.
//   - TestServicePushCommand_ProcessFails
//     Process polling returns FAILED → spinner reports the failure and the
//     command exits non-zero.
//   - TestServicePushCommand_SetupNotFoundInYaml_ForwardedToApi
//     Documents current behavior: a --setup value not present in the local
//     yaml is forwarded to the API unvalidated.
//
// ### Tier 2: polling, scope, archive-file-path
//
//   - TestServicePushCommand_PendingThenRunningThenFinished
//     Process status sequence PENDING→RUNNING→FINISHED across three polls.
//     Asserts the polling loop visits all of them.
//   - TestServicePushCommand_InvalidServiceIdErrors
//     A 404 with serviceStackNotFound on the service-stack endpoint exits
//     non-zero with a user-facing error (no panic).
//   - TestServicePushCommand_ArchiveFilePathTeesToFile
//     --archive-file-path writes a non-empty copy of the upload to disk via
//     the tee'd reader.
//   - TestServicePushCommand_ArchiveFilePathAlreadyExistsErrors
//     openPackageFile refuses to overwrite an existing archive file; the
//     push fails before any upload, and the existing file is untouched.
//   - TestServicePushCommand_ProjectFlagAndServiceByName
//     --project-id flag + non-UUID positional service arg drives the
//     by-name lookup path (UUID-as-id fails with serviceStackNotFound, then
//     GetServiceByName succeeds).
//   - TestServicePushCommand_ScopeFromSavedProjectId
//     With a previously persisted ScopeProjectId (set via SeedScopedLogin),
//     `zcli push <name>` resolves the project from saved scope without
//     --project-id.
//
// ## zcli service push — git archive paths (servicePushGit_integration_test.go)
//
// All tests in this file build a real `git init` repo in the test temp dir
// using runGit / gitInit / gitAddCommit helpers. They skip cleanly if `git`
// is not on PATH. Uploaded archive bytes are decompressed and walked via
// archiveEntries; entries are inspected by suffix.
//
//   - TestServicePushCommand_GitNotInitializedErrors
//     Working-dir without a .git directory + no --no-git → archive client
//     refuses with a "git init" message.
//   - TestServicePushCommand_GitZeroCommitsErrors
//     .git exists but zero commits → "at least one commit" error.
//   - TestServicePushCommand_GitArchive_CommittedFilesUploaded
//     Default git archive: committed main.go and zerops.yaml end up in the
//     uploaded gzipped tar.
//   - TestServicePushCommand_GitArchive_WorkspaceCleanIgnoresUncommitted
//     --workspace-state=clean omits uncommitted files in the working tree.
//   - TestServicePushCommand_GitArchive_WorkspaceStagedKeepsStagedDropsUnstaged
//     --workspace-state=staged includes the index, excludes pure working-
//     tree changes.
//   - TestServicePushCommand_GitArchive_WorkspaceAllIncludesUncommitted
//     Default --workspace-state=all captures committed + staged + unstaged
//     files.
//   - TestServicePushCommand_GitArchive_DeployGitFolderIncludesGitDir
//     --deploy-git-folder embeds the .git/ directory in the upload archive.
//
// ## Confirmed-bug reproductions (pushDeploy_bugs_test.go)
//
// Both tests are kept t.Skip'd so CI stays green; removing the Skip should
// produce the documented failure (or success, once the underlying bug is
// fixed) and lock the fix in.
//
//   - TestServicePushCommand_SetupFlagOverridesAutoMatch (skipped)
//     servicePush.go and serviceDeploy.go check the auto-match against the
//     service name BEFORE consulting the --setup flag. When the service name
//     happens to equal a setup name in zerops.yaml, the explicit --setup is
//     silently ignored. Reproduced from a real pipeline running
//     `--setup showcase-backend` on a service named "backend" with both
//     setups present. Fix: invert the branches so the flag wins.
//   - TestServicePushCommand_RunningProcessWithNullAppVersion_BUGPROBE (skipped)
//     servicePush.go's log-streaming callback dereferences
//     apiProcess.AppVersion.Id (push.go:235) and apiProcess.AppVersion.Build
//     (push.go:251) the first time a poll returns status=RUNNING. AppVersion
//     is a pointer in output.Process; if the API returns RUNNING before
//     AppVersion is populated, the CLI nil-derefs in the spinner goroutine
//     and the binary crashes with SIGSEGV. Confirmed by running the probe in
//     isolation. Fix: nil-guard before the deref sites.
package cmd
