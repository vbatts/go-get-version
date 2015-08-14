# go-get-version

version generator for use with go:generate

## install

```bash
go get github.com/vbatts/go-get-version
```

## usage

include the following in a place like `./version/gen.go`:

```go
package version
//go:generate go-get-version -package version -variable VERSION -output version.go
```
