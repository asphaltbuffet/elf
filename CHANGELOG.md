# Changelog

## [0.2.0] - 2024-07-11

### Changed

- **Breaking:** Replace `day`/`year` fields with filepath(s) ([`7be1488`](https://github.com/asphaltbuffet/elf/commit/7be14882cbf0af0ecf160edebfbed10d1975810f))
- Bump `go` from 1.21.x to 1.22.3 ([`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Change format of version output ([`cb3d03d`](https://github.com/asphaltbuffet/elf/commit/cb3d03d252e650fe6e71448e780a357fd03deaec))
- Refactor output formatting ([`#44`](https://github.com/asphaltbuffet/elf/pull/44))
- Refactor file handling ([`#22`](https://github.com/asphaltbuffet/elf/pull/22), [`4f37268`](https://github.com/asphaltbuffet/elf/commit/4f372685c679d23666dc799d8b576322ef2f5df6), [`4a0c656`](https://github.com/asphaltbuffet/elf/commit/4a0c656898e2525faa88d4df991a5876d2fa9d7f))
- Set default language as `Go` ([`0ac2ca9`](https://github.com/asphaltbuffet/elf/commit/0ac2ca92d30f2a707d901d27ad09f1c2a0c6cf86))
- Default logging level to `Info` instead of `Debug` ([`2ce88c9`](https://github.com/asphaltbuffet/elf/commit/2ce88c905a2dd9fe62d23335b1140596d5eda290))
- Refactor runtime tasks ([`#40`](https://github.com/asphaltbuffet/elf/pull/44), [`#60`](https://github.com/asphaltbuffet/elf/pull/60))
- Trim page data before writing to disk ([`#22`](https://github.com/asphaltbuffet/elf/pull/22))
- Refactor configuration into separate package ([`#44`](https://github.com/asphaltbuffet/elf/pull/44))
- Improve error handling around setting language ([`#25`](https://github.com/asphaltbuffet/elf/pull/25))
- Refactor flow of exercise loading ([`#24`](https://github.com/asphaltbuffet/elf/pull/24))
- Refactor `solver` and `tester` ([`#22`](https://github.com/asphaltbuffet/elf/pull/22), [`#23`](https://github.com/asphaltbuffet/elf/pull/23), [`#44`](https://github.com/asphaltbuffet/elf/pull/44))
- Decrease excessive logging ([`#22`](https://github.com/asphaltbuffet/elf/pull/22), [`3fcb971`](https://github.com/asphaltbuffet/elf/commit/3fcb9711cf10faeac4727e01574220cdea7f99a8,) [`a8047fe`](https://github.com/asphaltbuffet/elf/commit/a8047fea080339112c3542fd029da3e2ff4212b1,) [`#44`](https://github.com/asphaltbuffet/elf/pull/44), [`1c6c670`](https://github.com/asphaltbuffet/elf/commit/1c6c67086c90c0e30404e1fe43cfea380565304e))
- Bump `charmbracelet/lipgloss` from 0.9.1 to 0.11.0 ([`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Bump `go-resty/resty` from 2.10.0 to 2.13.1 ([`11102e5`](https://github.com/asphaltbuffet/elf/commit/11102e5af4717b6ac7ea1b3bc33a8947b7778488), [`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Bump `lmittmann/tint` from 1.0.2 to 1.0.4 ([`165ad67`](https://github.com/asphaltbuffet/elf/commit/165ad67d5f01afe1f7071229a0a74b2c9f251900), [`11102e5`](https://github.com/asphaltbuffet/elf/commit/11102e5af4717b6ac7ea1b3bc33a8947b7778488))
- Bump `spf13/afero` from 1.10.0 to 1.11.0 ([`0fd01c0`](https://github.com/asphaltbuffet/elf/commit/0fd01c04a74dbcc31ab7ffb8f96389b4e808cf80), [`11102e5`](https://github.com/asphaltbuffet/elf/commit/11102e5af4717b6ac7ea1b3bc33a8947b7778488), [`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Bump `spf13/cobra` from 1.8.0 to 1.8.1 ([`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Bump `spf13/viper` 1.16.0 to 1.19.0 ([`0fd01c0`](https://github.com/asphaltbuffet/elf/commit/0fd01c04a74dbcc31ab7ffb8f96389b4e808cf80), [`11102e5`](https://github.com/asphaltbuffet/elf/commit/11102e5af4717b6ac7ea1b3bc33a8947b7778488), [`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Bump `golang.org/x/net` from 0.17.0 to 0.27.0 ([`4742a52`](https://github.com/asphaltbuffet/elf/commit/4742a52189c651da7f4a2b4563c5d91b6fb0124b), [`11102e5`](https://github.com/asphaltbuffet/elf/commit/11102e5af4717b6ac7ea1b3bc33a8947b7778488), [`34b0564`](https://github.com/asphaltbuffet/elf/commit/34b0564db933453c41ae5f15b0637cb19dfb23ef), [`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))
- Bump `golang.org/x/text` from 0.13.0 to 0.16.0 ([`4742a52`](https://github.com/asphaltbuffet/elf/commit/4742a52189c651da7f4a2b4563c5d91b6fb0124b), [`2c66033`](https://github.com/asphaltbuffet/elf/commit/2c660338c4083af3c465e5196dd8925ab1c3dfd4))

### Added

- Add linux packages to release artifacts ([`0748146`](https://github.com/asphaltbuffet/elf/commit/07481468ec3a9a6724333d1bf741bcc5be6196ca))
- Add `benchmark` command ([`#44`](https://github.com/asphaltbuffet/elf/pull/44), [`a60849a`](https://github.com/asphaltbuffet/elf/commit/a60849a930fd336d85d4471525109bad59432e81))
- Add `graph` subcommand ([`#44`](https://github.com/asphaltbuffet/elf/pull/44), [`728398f`](https://github.com/asphaltbuffet/elf/commit/728398fb0e6eda91f0eb30417edd275012885c64), [`f4408ec`](https://github.com/asphaltbuffet/elf/commit/f4408ec6feff54393e94ce1b1884ff3b7cea2e1a))
- Add flag to skip tests when running `solve` command ([`#24`](https://github.com/asphaltbuffet/elf/pull/24))
- Add flag to specify config file ([`74d81ee`](https://github.com/asphaltbuffet/elf/commit/74d81ee4a7c2d3a8d9f291ce1faa5baa40146d73))
- Add flag to specify runtime input file ([`2a4a5f8`](https://github.com/asphaltbuffet/elf/commit/2a4a5f8b9e25d545c4f3b0400e7fc92977b39e1c), [`728398f`](https://github.com/asphaltbuffet/elf/commit/728398fb0e6eda91f0eb30417edd275012885c64))
- Add config setting for exercise data filename ([`1beee4e`](https://github.com/asphaltbuffet/elf/commit/1beee4e08c6e6800d010c1ca1ed126ca54b8855f))
- Auto-generate documentation and completions ([`16605de`](https://github.com/asphaltbuffet/elf/commit/16605deeb81899b9e9eb790a1a5339c0e05518cd), [`abb5b57`](https://github.com/asphaltbuffet/elf/commit/abb5b57c093c47e3bc2e8e669dd4aeac9159ba1c))
- Add project Taskfile ([`#20`](https://github.com/asphaltbuffet/elf/pull/20), [`#22`](https://github.com/asphaltbuffet/elf/pull/22))
- Add license ([`4762dc5`](https://github.com/asphaltbuffet/elf/commit/4762dc524a798dd470bb5454df158311e5b1202f))

### Removed

- Remove `Makefile` and `justfile` ([`#53`](https://github.com/asphaltbuffet/elf/pull/53))
- Remove `Exercise` package ([`#44`](https://github.com/asphaltbuffet/elf/pull/44))

### Fixed

- Check if files exist before writing ([`#44`](https://github.com/asphaltbuffet/elf/pull/44))
- Prevent writing error output to files ([`#22`](https://github.com/asphaltbuffet/elf/pull/22))
- Fix python lib path failure ([`#54`](https://github.com/asphaltbuffet/elf/pull/54))
- Fix exercise readme link in advent year readme template ([`#44`](https://github.com/asphaltbuffet/elf/pull/44))
- Fix incorrect handling of `'` and `-` in titles ([`#21)`](https://github.com/asphaltbuffet/elf/pull/21))
- Detect invalid configuration when downloading advent data ([`#22`](https://github.com/asphaltbuffet/elf/pull/22))
- Walk exercise directory to allow for deeper nesting ([`#22`](https://github.com/asphaltbuffet/elf/pull/22))

## [0.1.0] - 2023-11-12

_🌱 Initial release_

[0.2.0]: https://github.com/asphaltbuffet/elf/releases/tag/v0.2.0
[0.1.0]: https://github.com/asphaltbuffet/elf/releases/tag/v0.1.0
