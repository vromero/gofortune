<p align="center">
  <img alt="GoFortune Logo" src="https://openclipart.org/image/300px/svg_to_png/181849/fortune-cookie.png" height="140" />
  <h3 align="center">GoFortune</h3>
  <p align="center">Print a random, hopefully interesting, adage.</p>
  <p align="center">
    <a href="https://github.com/vromero/gofortune/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/vromero/gofortune.svg?style=flat-square"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-Apache%202-blue.svg?style=flat-square"></a>
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
gofortune
```

Useful flags:
- `-s` short fortunes only, `-l` long only, `-n N` threshold (default 160)
- `-m PATTERN` print all fortunes matching a regular expression, `-i` case-insensitive
- `-o` pick from offensive fortunes only, `-a` all maxims
- `-c` show the cookie file a fortune came from
- `-f` print the list of candidate files and their probabilities
- `-e` weight every file equally regardless of size
- `-w` pause after printing, scaling with the length of the fortune

Provide one or more paths (optionally preceded by `N%` to weight them) to
override the default `/usr/share/games/fortunes` location:
```bash
gofortune 30% /path/to/my/fortunes 70% /path/to/other/fortunes
```

### Strfile
Create a random access index file for storing strings:
```bash
gofortune strfile <source_file> [data_file]
```

Example:
```bash
# Create an index from fortunes.txt and save it as fortunes.txt.dat
gofortune strfile fortunes.txt
```

## I18n (Internationalization)

GoFortune supports multiple languages. When the `LANG` environment variable is set, the tool will attempt to find fortunes in the corresponding directory (e.g., `/usr/share/games/fortunes/es`).

## Contributing

This project adheres to the [Contributor Covenant](CODE_OF_CONDUCT.md) code of conduct. Please refer to our [contributing guidelines](CONTRIBUTING.md) for more information.
