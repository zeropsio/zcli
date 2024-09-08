# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.24] - 2024-09-08

### Changed
- Removed the need to provide the --projectId flag for service scope commands when the --serviceId flag is provided.

## [v1.0.22] - 2024-08-05

### Added
- install scripts for Linux, macOS and Windows
- check if `resolvectl` is available for `vpn` commands

### Fixed
- commands will now exit with exit code 1 for all error types

## [v1.0.20] - 2024-07-02

### Added
- `service deploy` command now reacts to the `.deployignore` file located in the `--workingDir` directory, learn more about `.deployignore` in the docs

## [v1.0.19] - 2023-06-06

### Fixed
- `go get` would fail due to the `non–ascii.txt` file in one of the test cases. File is now created ad-hoc during the test and then cleaned up.

## [v0.12.19] - 2023-02-24

### Fixed
- `service log` and `vpn start` were ignoring env and config values for `limit` and `mtu` parameters respectively

### Changed
- released binaries do not include debug tables and should be about 25% to 30% smaller
- `push` command now correctly pushes all files if called from a repository utilizing git submodules

## [v0.12.18] - 2023-01-21

### Added
- `zcli init` command

## [v0.12.17] - 2023-01-10

### Changed
- use `ping -6` when `ping6` is not available

## [v0.12.16] - 2022-11-04

### Fixed
- fix OSX vpn DNS setup 

## [v0.12.15] - 2022-11-01

### Fixed
- fix OSX vpn DNS setup

## [v0.12.14] - 2022-10-12

### Fixed
- support absolute path for import script
- fix missing clientId in get logs on appVersion and container

## [v0.12.13] - 2022-10-06

### Fixed
- remove printing null or empty error meta

## [v0.12.12] - 2022-10-06

### Fixed
- project import on windows

## [v0.12.11] - 2022-09-30

### Fixed
- better VPN Darwin support with DHCP setup

## [v0.12.10] - 2022-09-27

### Fixed
- replace old `zerops-io` repository to new `zeropsio` repository name

## [v0.12.9] - 2022-09-27

### Added
- `--follow` flag for `zcli service logs` command to receive continuous stream of logs

## [v0.12.8] - 2022-09-26

### Added
- `vpn start [ --preferredPort PORT_RANGE ]` parameter

### Fixed
- windows vpn setup 
- linux vpn setup 
- darwin vpn setup 

## [v0.12.7] - 2022-09-08

### Added
- `--region` flag with `REGION` env option to the `zcli bucket s3` `create` and `delete` commands
 
## [v0.12.6] - 2022-09-01

### Fixed
- `zerops.yml` file is optional for certain service types in `deploy` command

### Added
- Validation of `zerops.yml` file into `deploy` command

## [v0.12.5] - 2022-08-30

### Fixed
- `zcli deploy` would not work on Windows when certain formats of paths were passed as parameters.

## [v0.12.4] - 2022-08-24

### Fixed
- Return correct error messages when project not found by name or ID.

### Added
- New set of S3 management `bucket` commands with ability to `create` and `delete` buckets
  - via `Zerops API`:
    - `zcli bucket zerops create projectNameOrId serviceName bucketName [flags]`
    - `zcli bucket zerops delete projectNameOrId serviceName bucketName [flags]`
  - via `S3 API`:
    - `zcli bucket s3 create serviceName bucketName [flags]`
    - `zcli bucket s3 delete serviceName bucketName [flags]`

## [v0.12.3] - 2022-08-22

### Added
- PersistentKeepalive for windows VPN clients

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
