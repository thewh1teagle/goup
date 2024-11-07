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
go build -ldflags="-X 'main.Tag=v0.0.0'" main.go
./main.exe
```