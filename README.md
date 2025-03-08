# Brux

[![Lint Markdown](https://github.com/ashishb/brux/actions/workflows/lint-markdown.yaml/badge.svg)](https://github.com/ashishb/brux/actions/workflows/lint-markdown.yaml)
[![Lint YAML](https://github.com/ashishb/brux/actions/workflows/lint-yaml.yaml/badge.svg)](https://github.com/ashishb/brux/actions/workflows/lint-yaml.yaml)
[![Lint GitHub Actions](https://github.com/ashishb/brux/actions/workflows/lint-github-actions.yaml/badge.svg)](https://github.com/ashishb/brux/actions/workflows/lint-github-actions.yaml)

[![Lint Go](https://github.com/ashishb/brux/actions/workflows/lint-go.yaml/badge.svg)](https://github.com/ashishb/brux/actions/workflows/lint-go.yaml)
[![Validate Go code formatting](https://github.com/ashishb/brux/actions/workflows/format-go.yaml/badge.svg)](https://github.com/ashishb/brux/actions/workflows/format-go.yaml)

Brux is a CLI tool written in Go for executing [Bruno](https://github.com/usebruno/bruno)'s [Bru files](https://github.com/brulang/bru-lang).

Features

- [x] Parse .bru file
- [x] Run .bru file
- [x] Support enviroments
- [x] Support .env variables
- [x] Support saving output
- [x] Pretty print JSON
- [ ] Add ability to run against multiple environments and compare the results

## Install

```bash
$ go install github.com/ashishb/brux/src/brux/cmd/brux@latest
...
```

Or run it directly

```bash
$ go run github.com/ashishb/brux/src/brux/cmd/brux@latest --help
Usage:
  brux [flags]
  brux [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Run a Bru file

Flags:
  -h, --help   help for brux
```

## Usage

Consider a sample `example.bru` file

```bru
meta {
  name: Send request to example.com
  type: http
  seq: 1
}

get {
  url: http://example.com/
  body: json
  auth: none
}

headers {
  Content-Type: application/json
}
```

You can run it as

```bash
$ go run github.com/ashishb/brux/src/brux/cmd/brux@latest run example.bru
...
```
