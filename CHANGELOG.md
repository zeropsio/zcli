# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2022-05-12

### Added
- New command `zcli region list`, which lists available regions to the user
- Hint user the possibility to change the region when auth error occurs
- Support id, which is printed to stdin on `internal server error`
- Fix an error with incorrect certificate server name