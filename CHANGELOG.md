# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [0.0.1] - 2024-09-04

First release with bare minimum set of functionalities to make it usable for
basic needs.

### Added

- Proxy all http requests to lambdas, routing based on host, working both
  locally and on AWS
- Acceptance tests working on local lambdas to ensure the proxy based on host
  works and to check if concurrent requests are handled correctly (local lambdas
  do not accept concurrent requests)
- Github Action workflow to build and check the code
- GitHub Action workflow to release on Docker Hub when commit is tagged
- README.md explaining what's the project and how to use it

### Changed

### Fixed
