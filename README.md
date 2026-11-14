# gofile

A `file(1)`-like utility written in Go. It identifies Mach-O, ELF, and PE
binaries and reports their format, word size, file type, and CPU
architecture, using only the standard library's `debug` packages.

## Install

Requires Go 1.26 or later.

```bash
go install -ldflags "-s -w" github.com/mikemadden42/gofile@latest
```

## Usage

```
gofile [-v] file...
```

Each file gets one line on stdout:

```console
$ gofile gofile libf.dylib /usr/bin/true hello-linux-pie hello-windows-386.exe
gofile: Mach-O 64-bit executable arm64
libf.dylib: Mach-O 64-bit dynamically linked shared library arm64
/usr/bin/true: Mach-O universal binary [amd64 arm64e arm64]
hello-linux-pie: ELF 64-bit pie executable amd64
hello-windows-386.exe: PE 32-bit executable 386
```

Architectures use Go's `GOARCH` names (`amd64`, `arm64`, `386`, `ppc64le`, …),
with `arm64e` for Apple's pointer-authentication slices.

Files that can't be read or aren't a recognized binary are reported on stderr:

```console
$ gofile missing /etc/hosts
missing: stat missing: no such file or directory
/etc/hosts: unrecognized format
```

Pass `-v` to see why each format didn't match:

```console
$ gofile -v /etc/hosts
/etc/hosts: not Mach-O: invalid magic number in record at byte 0x0
/etc/hosts: not Mach-O universal: invalid magic number in record at byte 0x0
/etc/hosts: not ELF: bad magic number '[35 35 10 35]' in record at byte 0x0
/etc/hosts: not PE: unrecognized PE machine: 0x2323
/etc/hosts: unrecognized format
```

### Exit status

| Code | Meaning |
| ---- | ------- |
| 0 | Every file was identified |
| 1 | At least one file couldn't be read or wasn't recognized |
| 2 | No files were given |

## Development

```bash
go test ./...
make check    # golangci-lint
make build    # goreleaser snapshot build into dist/
```

Based on the example program in
[this post on jvt.me](https://www.jvt.me/posts/2023/05/15/go-parse-binary-architecture/).
