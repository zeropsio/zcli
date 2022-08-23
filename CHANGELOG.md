# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [v0.12.3] - 2022-08-23

### Fixed
- Return correct error messages when project not found by name or ID.

### Added
- New set of S3 management `bucket` commands with ability to `create` and `delete` buckets
  - via `Zerops API`:
    - `zcli bucket zerops create projectNameOrId serviceName bucketName [flags]`
    - `zcli bucket zerops delete projectNameOrId serviceName bucketName [flags]`
  - via `S3 API`.
    - `zcli bucket s3 create serviceName bucketName [flags]`
    - `zcli bucket s3 delete serviceName bucketName [flags]`.

## [v0.12.2] - 2022-08-16

### Fixed
- Inherit the `PATH` variable from the user on `daemon install` on `darwin` platform.

## [v0.12.1] - 2022-08-09

### Fixed
- Added missing default URL for region list command.

## [v0.12.0] - 2022-08-08

### Changed
- Updated protobufs to the latest version (**!!!breaking change!!! previous zCLI versions are not compatible and will not work**).
- Updated `protoc-gen` from GitHub to `protoc-gen-go` and `protoc-gen-go-grpc` from GoLang.org.

## [v0.11.4] - 2022-07-26
- Enable lowercase formatTemplate values, fix length of timestamps.

## [v0.11.3] - 2022-07-25
- Update commands descriptions.

## [v0.11.2] - 2022-07-14
- Accept lowercase values for service log flags.

## [v0.11.1] - 2022-07-04
- Improve error messages.

## [v0.11.0] - 2022-06-30
- Add service log command.

## [v0.10.2] - 2022-06-17
- Hide internal flags from help, hide completion command.

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
