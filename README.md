# goup

Add self update capability for go program

## Features

-   Update the binary from Github releases
-   Simple interface (2 functions)
-   Support macOS / Linux / Windows

## Build

```console
go build -ldflags="-X 'main.Tag=v0.1.0'" main.go
```

## Test

```console
echo test > goup_windows_x86_64.exe
gh release upload v0.1.0 goup_windows_x86_64.exe --clobber
go build -ldflags="-X 'main.Tag=v0.0.0'" cmd/main.go
./main.exe
```

## Release

```console
$outdir="bin"
$name="goup"
Remove-Item -Force -Recurse $outdir ; mkdir $outdir
$LDFLAGS="-X 'main.Version=$(git describe --tags --abbrev=0)'"
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o "$outdir/${name}_darwin_x86_64" -ldflags $LDFLAGS cmd/main.go
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o "$outdir/${name}_darwin_arm64" -ldflags $LDFLAGS cmd/main.go
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o "$outdir/${name}_windows_x86_64.exe" -ldflags $LDFLAGS cmd/main.go
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o "$outdir/${name}_linux_x86_64" -ldflags $LDFLAGS cmd/main.go
```
