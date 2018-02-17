# url - command-line URL parser

`url` parses its arguments as URLs and prints a structured representation (JSON or go templates)

- [Get it](#get-it)
- [Use it](#use-it)
    - [JSON output](#json-output)
    - [Template output](#template-output)
- [Example](#example)
- [Comments](https://github.com/sgreben/url/issues/1)


## Get it

Using go get:

```bash
go get -u github.com/sgreben/url/cmd/url
```

Or [download the binary](https://github.com/sgreben/url/releases/latest) from the releases page.

Also available as a [docker image](https://quay.io/repository/sergey_grebenshchikov/url?tab=tags):

```bash
docker pull quay.io/sergey_grebenshchikov/url
```

## Use it

`url` reads URLs from CLI arguments and writes to stdout.

```text
Usage of url:
  -r    alias for -resolve
  -resolve
        resolve ../ in URLs
  -t string
        alias for -template
  -template string
        go template output
  -version
        print version and exit
```

### JSON output

The default output format is JSON, one object per line:

```bash
$ url https://github.com/sgreben/url/cmd/url
```

```json
{"scheme":"https","hostname":"github.com","host":"github.com","path":"/sgreben/url/cmd/url","pathComponents":["sgreben","url","cmd","url"],"query":{},"port":"","fragment":""}
```

### Template output

You can specify an output template using the `-template` parameter and [go template](https://golang.org/pkg/text/template) syntax:

```bash
$ url -t .Hostname https://github.com/sgreben/url/cmd/url
```

```text
github.com
```

The fields available to the template are specified in the [`flatURL` struct](cmd/url/main.go#L15).

## Comments

Feel free to [leave a comment](https://github.com/sgreben/url/issues/1) or create an issue.