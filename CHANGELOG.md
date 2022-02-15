# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

- Use Otto instead of Goja
- Conform tests (format/style) with Golang standard
- Get accessors for PACman information

## [0.0.11] - 2022-02-14

### Added

- Added ability to handle PAC filepath with or without URI, *nix or Windows, local or remote.
- Upgrade golang-ci version (CI pipeline)
- standardized Go version to 1.16

## [0.0.10] - 2021-10-27
### Added
- Added ability to handle `localhost` pac proxy URI synonyms (`127.0.0.1`, and `0.0.0.0`).

## [0.0.9] - 2021-10-13
### Changed
- Upgraded dependencies, removed `PACMAN_LOG_LEVEL`. Sypl supports logging filtering, and max level fine-control.

## [0.0.8] - 2021-10-10
### Added
- Added ability to specify credentials (PAC URI, PAC proxies URI) via env var.

## [0.0.7] - 2021-10-09
### Added
- Prints redacted URI instead of plain one.

## [0.0.6] - 2021-10-09
### Added
- Added the ability to request PAC content from a protected remote (HTTP/HTTPS) server.

## [0.0.5] - 2021-10-08
### Added
- Added Sypl.
  - `PACMAN_LOG_LEVEL` env var controls the logging level.
- Added `GetXYZ` accessors for private fields.
- Added the ability to set credentials for PAC proxies.
- Added more tests, and covered more cases.
- Added more validation, and validators.

### Changed
- Started breaking down code into packages (`internal`, `pkg`).
- Now the type of the PAC proxy is an enum (`mode`).
- Improved some of the matching algol using pre-compiled (optimized) regex.
- Fixed some of the naming inconsistencies.
- Improved documentation.

## [0.0.4] - 2021-10-07

- Refresh registry.

## [0.0.3] - 2021-10-07
### Added
- `parser.Source` now returns the source of the PAC content.
- Added more tests.

### Changed
- `parser.Source` renamed to `parser.Content`.

## [0.0.2] - 2021-10-07
### Added
- Universal loader (text, file, remote).
- `parser.Source` returns the loaded PAC content.

## [0.0.1] - 2021-10-06

First release.

### Checklist

- [x] CI Pipeline:
  - [x] Lint
  - [x] Tests
- [x] Documentation:
  - [ ] Package's documentation (`doc.go`)
  - [ ] Meaningful code comments, and symbol names (`const`, `var`, `func`)
  - [x] `GoDoc` server tested
  - [ ] `README.md`
  - [x] `LICENSE`
    - [x] Files has LICENSE in the header
  - [x] Useful `CHANGELOG.md`
  - [x] `CONTRIBUTION.md`
- Automation:
  - [x] `Makefile`
- Testing:
  - [ ] Coverage 80%+
  - [x] Unit test
  - [x] Real testing
- Examples:
  - [x] Example's test file
- Errors:
  - [ ] Consistent, and standardized errors (powered by `CustomError`)
- Logging:
  - [ ] Consistent, and standardized logging (powered by `Sypl`)
  - [ ] Output to `stdout`
  - [ ] Output to `stderr`
  - [ ] Output to file
