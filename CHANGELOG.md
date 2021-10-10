# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

- Use Otto instead of Goja
- Conform tests (format/style) with Goland standard
- Move generated SOCKS (`socks.go`) client to internal

## [0.0.8] - 2021-10-10

- [x] Added ability to specify credentials (PAC URI, PAC proxies URI) via env var.

## [0.0.7] - 2021-10-09

- [x] Prints redact URI instead of plain one.

## [0.0.6] - 2021-10-09

- [x] Added the ability to request PAC content from a protected remote (HTTP/HTTPS) server.

## [0.0.5] - 2021-10-08

- [x] Added Sypl.
  - [x] `PACMAN_LOG_LEVEL` env var controls the logging level.
- [x] Added `GetXYZ` accessors for private fields.
- [x] Added the ability to set credentials for PAC proxies.
- [x] Improved documentation.
- [x] Added more tests, and covered more cases.
- [x] Added more validation, and validators.
- [x] Fixed some of the naming inconsistencies.
- [x] Now the type of the PAC proxy is an enum (`mode`).
- [x] Improved some of the matching algol using pre-compiled (optmized) regex.
- [x] Started breaking down code into packages (`internal`, `pkg`).

## [0.0.4] - 2021-10-07

- [x] Refresh registry.

## [0.0.3] - 2021-10-07

### Added

- [x] `parser.Source` now returns the source of the PAC content.
- [x] Added more tests.

### Changed

- `parser.Source` renamed to `parser.Content`.

## [0.0.2] - 2021-10-07

### Added

- [x] Universal loader (text, file, remote).
- [x] `parser.Source` returns the loaded PAC content.

## [0.0.1] - 2021-10-06

### Added/Removed/Changed

- [x] Removed anything non-PAC parsing related.
- [x] Modernized code.
- [x] Made it work properly/fixed bugs/bad code.
- [x] Fixed/added more tests.

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
