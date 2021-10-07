# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

- Use Otto instead of Goja
- Conform tests (format/style) with Goland standard
- Cover with more tests
- Improve general code quality

## [0.0.1] - 2021-10-06

### Added/Removed/Changed

- [x] Removed anything non-PAC parsing related
- [x] Modernized code
- [x] Made it work properly/fixed bugs/bad code
- [x] Fixed/added more tests

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
