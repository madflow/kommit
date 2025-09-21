# Changelog

## [0.6.0](https://github.com/madflow/kommit/compare/v0.5.6...v0.6.0) (2025-09-21)


### Features

* add pull request support to kommit tool ([73b8893](https://github.com/madflow/kommit/commit/73b88936ad12aee2245d96a6b8322d8605ccdf7f))


### Bug Fixes

* Arrr, I be fixing the treasure map parser! ([10c4ad8](https://github.com/madflow/kommit/commit/10c4ad8b20afcfb87aff9693a8d51ea013d7c59d))

## [0.5.6](https://github.com/madflow/kommit/compare/v0.5.5...v0.5.6) (2025-09-20)


### Bug Fixes

* **config:** update default rule format ([0e66166](https://github.com/madflow/kommit/commit/0e661665a4790bef808e4072d9e0f7a686853d77))
* **quarantine:** remove macOS quarantine attribute during release ([08d9288](https://github.com/madflow/kommit/commit/08d9288f1dbc6224258db1b5b6d5443df158dad4))

## [0.5.5](https://github.com/madflow/kommit/compare/v0.5.4...v0.5.5) (2025-09-08)


### Bug Fixes

* **deps:** update module github.com/spf13/viper to v1.21.0 ([#37](https://github.com/madflow/kommit/issues/37)) ([9761661](https://github.com/madflow/kommit/commit/9761661d190eff780d3850e27daef170464f5fb8))

## [0.5.4](https://github.com/madflow/kommit/compare/v0.5.3...v0.5.4) (2025-09-07)


### Bug Fixes

* **goreleaser:** remove deprecated conflicts_with formula ([976e0cd](https://github.com/madflow/kommit/commit/976e0cd2bd5647fc8b2ba51fee35ca6671a71f52))

## [0.5.3](https://github.com/madflow/kommit/compare/v0.5.2...v0.5.3) (2025-09-01)


### Bug Fixes

* **deps:** update module github.com/spf13/cobra to v1.10.1 ([#31](https://github.com/madflow/kommit/issues/31)) ([b2912ee](https://github.com/madflow/kommit/commit/b2912ee42b587d1327b1a0a86ef419fb76aab1cd))

## [0.5.2](https://github.com/madflow/kommit/compare/v0.5.1...v0.5.2) (2025-09-01)


### Bug Fixes

* **deps:** update module github.com/spf13/cobra to v1.10.0 ([#29](https://github.com/madflow/kommit/issues/29)) ([a891170](https://github.com/madflow/kommit/commit/a891170f185b2231cccef420a006dab6e32b4665))

## [0.5.1](https://github.com/madflow/kommit/compare/v0.5.0...v0.5.1) (2025-07-29)


### Bug Fixes

* **cmd/branch.go:** sanitize branch names to conform to git standards ([f1240d5](https://github.com/madflow/kommit/commit/f1240d5acbceb404063a8a965ea32c1123c120f2))

## [0.5.0](https://github.com/madflow/kommit/compare/v0.4.5...v0.5.0) (2025-07-29)


### Features

* **cmd:** add branch command for generating branch names based on changes ([c77edb0](https://github.com/madflow/kommit/commit/c77edb02e9e8a690c4c0177c32df0985912ad909))


### Bug Fixes

* **branch:** remove redundant import and simplify log message ([93d61ff](https://github.com/madflow/kommit/commit/93d61ff0902ce31b1a631f845eb477d0942993c5))

## [0.4.5](https://github.com/madflow/kommit/compare/v0.4.4...v0.4.5) (2025-07-05)


### Bug Fixes

* **.kommit.yaml:** correct commit message guidance ([898fda3](https://github.com/madflow/kommit/commit/898fda3f4bc11afee1a5742e177d54d064f8d8c3))
* correct branch display name ([fadf0c9](https://github.com/madflow/kommit/commit/fadf0c95d2e1e8a0cab27598b1421593f6942dd5))
* update branch name rules and commit message guidelines ([13a98eb](https://github.com/madflow/kommit/commit/13a98eb0d53b46b2f5432bc033a736dba5b4004b))

## [0.4.4](https://github.com/madflow/kommit/compare/v0.4.3...v0.4.4) (2025-06-30)


### Bug Fixes

* **config:** update message rules ([75738d9](https://github.com/madflow/kommit/commit/75738d996833de203682f6ab887f8f64b510f0fa))

## [0.4.3](https://github.com/madflow/kommit/compare/v0.4.2...v0.4.3) (2025-06-27)


### Bug Fixes

* **workflow:** update GoReleaser action version and configuration ([4a97baf](https://github.com/madflow/kommit/commit/4a97baf1cbb001505f867670a66fa7ef7e172f5f))

## [0.4.2](https://github.com/madflow/kommit/compare/v0.4.1...v0.4.2) (2025-06-27)


### Bug Fixes

* **release:** fix goreleaser integration ([584c4c7](https://github.com/madflow/kommit/commit/584c4c7e0ed7289ba24e6234c6142fb8582ed04e))

## [0.4.1](https://github.com/madflow/kommit/compare/v0.4.0...v0.4.1) (2025-06-27)


### Bug Fixes

* **goreleaser:** remove unnecessary changelog footer ([a8287a0](https://github.com/madflow/kommit/commit/a8287a04d6c7b22ccc4b880cef395521881d44ec))

## [0.4.0](https://github.com/madflow/kommit/compare/v0.3.0...v0.4.0) (2025-06-27)


### Features

* update goreleaser configuration to include full changelog URL and custom footer ([800b334](https://github.com/madflow/kommit/commit/800b334728bfa4b52deaa22672a2ae594742536c))

## [0.3.0](https://github.com/madflow/kommit/compare/v0.2.0...v0.3.0) (2025-06-27)


### Features

* update usage instructions for basic and advanced options ([1015ddb](https://github.com/madflow/kommit/commit/1015ddbf82a9ad3d0867045f9cbb6a3bdedb6ed7))

## [0.2.0](https://github.com/madflow/kommit/compare/v0.1.0...v0.2.0) (2025-06-16)


### Features

* **cmd/root.go:** add commit message editing functionality ([f9fda06](https://github.com/madflow/kommit/commit/f9fda06ee33cb1b2d4048cd88e239d40fbd18f1a))


### Bug Fixes

* **cmd/root.go:** allow multiple edits ([c60c00e](https://github.com/madflow/kommit/commit/c60c00edc6e8000a31f39cc8ce54913798e74f40))
* update .kommit.yaml to remove reference issues and pull requests requirement ([551ee84](https://github.com/madflow/kommit/commit/551ee840877349e47ce56816910028e790452ec4))

## [0.1.0](https://github.com/madflow/kommit/compare/v0.0.2...v0.1.0) (2025-06-16)


### Features

* create release-please.yml ([a305a22](https://github.com/madflow/kommit/commit/a305a22ca3d0d376974010ae7ce3f96095a61826))
