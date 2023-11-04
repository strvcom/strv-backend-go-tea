# Changelog
How to release a new version:
- Update this file with the latest version.
- Manually release new version.

## [Unreleased]
### Added
- `gen_id` now includes a comment at the top of the generated file to warn developers that the file is, in fact, generated.

## [0.4.0] - 2023-03-27
### Added
- Support for string IDs.

### Removed
- Useless marshalling/unmarshalling of custom IDs.

### Changed
- Updated Go version to 1.20.

## [0.3.1] - 2023-02-07
### Added
- Repo init command.

## [0.3.0] - 2022-10-30
### Added
- Scan method for uuid IDs.

### Fixed
- IDs generated file format.

## [0.2.2] - 2022-10-26
### Fixed
- Openapi works with updated dependencies.

## [0.2.1] - 2022-10-23
### Fixed
- Command for ID generating: fixed imports.

## [0.2.0] - 2022-10-15
### Added
- Command for ID generating: support for uuid type.

## [0.1.1] - 2022-09-22
### Added
- Added openapi support for extending RequestBody.

## [0.1.0] - 2022-08-01
### Added
- Added Changelog.

[Unreleased]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.2.0..v0.2.1
[0.2.0]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.1.1..v0.2.0
[0.1.1]: https://github.com/strvcom/strv-backend-go-tea/compare/v0.1.0..v0.1.1
[0.1.0]: https://github.com/strvcom/strv-backend-go-tea/releases/tag/v0.1.0
