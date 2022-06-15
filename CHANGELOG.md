# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.10.1] - 2022-06-15
- Add missing --source flag.

## [v0.10.0] - 2022-06-13
- Enable usage of project ID instead of project name.

## [v0.9.1] - 2022-06-09

### Fixed
- Fix corrupted archives from `push` and `deploy` commands on Windows platform.

## [v0.9.0] - 2022-06-01

### Added
- New flag `deployGitFolder` for `push` command which packs `.git` folder along other files for the `build` phase.

### Changed
- Archives stored by `push` and `deploy` commands now use `tar.gz` format instead of `zip`.
- Flag `zipFilePath` was renamed to `archiveFilePath`

## [v0.8.2] - 2022-05-31

### Added
- New command `zcli region list`, which lists available regions to the user.
- Hint user the possibility to change the region when auth error occurs.
- Support id, which is printed to stdin on `internal server error`.
- Fix an error with incorrect certificate server name.
- New commands `zcli project` and `zcli service`, both with subcommands `import`, `start`, `stop` and `delete` for full project and services management.
- Increase timeout values.
