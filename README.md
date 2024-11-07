# goup

Add self update capability for go program


## Build

```console
go run -ldflags="-X 'main.Tag=v0.1.0'" main.go
```

## Test

```console
echo test > goup_darwin_aarch64
gh release upload v0.1.0 goup_darwin_aarch64 --clobber
```