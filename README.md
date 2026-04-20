<p align="center">
  <img alt="GoFortune Logo" src="https://openclipart.org/image/300px/svg_to_png/181849/fortune-cookie.png" height="140" />
  <h3 align="center">GoFortune</h3>
  <p align="center">Print a random, hopefully interesting, adage.</p>
  <p align="center">
    <a href="https://github.com/vromero/gofortune/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/vromero/gofortune.svg?style=flat-square"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-Apache%202-blue.svg?style=flat-square"></a>
    <a href="https://github.com/goreleaser"><img alt="Powered By: GoReleaser" src="https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square"></a>
  </p>
</p>

---

GoFortune is a Go implementation of the classic Unix `fortune` and `strfile` utilities. It aims for compatibility with the `fortune-mod` 1.0 storage format.

## Installation

You can install GoFortune using `go install`:

```bash
go install github.com/vromero/gofortune@latest
```

## Usage

### Fortune
Print a random adage:
```bash
gofortune fortune
```

### Strfile
Create a random access index file for storing strings:
```bash
gofortune strfile <source_file> [data_file]
```

### Get

Download and install a fortune cookie collection from a GitHub repository. The repository should contain text files where each fortune is separated by a percent sign (`%`).

```bash
gofortune get <repository_url>
```

## I18n (Internationalization)

GoFortune supports multiple languages. When the `LANG` environment variable is set, the tool will attempt to find fortunes in the corresponding directory (e.g., `/usr/share/games/fortunes/es`).

## Contributing

This project adheres to the [Contributor Covenant](CODE_OF_CONDUCT.md) code of conduct. Please refer to our [contributing guidelines](CONTRIBUTING.md) for more information.
