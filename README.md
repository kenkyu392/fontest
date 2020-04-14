# fontest

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-00ADD8?logo=go)](https://pkg.go.dev/github.com/kenkyu392/fontest)
[![go report card](https://goreportcard.com/badge/github.com/kenkyu392/fontest)](https://goreportcard.com/report/github.com/kenkyu392/fontest)
[![license](https://img.shields.io/github/license/kenkyu392/fontest.svg)](LICENSE)

Fontest is a CLI tool for checking the characters included in font files.

## Installation

```
go get -u github.com/kenkyu392/fontest/cmd/fontest
```

## Usage

```console
$ fontest -file=characters.txt ./NotoSans-Regular.ttf ./NotoSansJP-Regular.ttf
_output/__NotoSans-Regular.png
_output/__NotoSansJP-Regular.png
_output/2020-01-01_000000.csv
```

## License

[MIT](LICENSE)
