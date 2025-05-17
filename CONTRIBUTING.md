# Contributing

### Requirements

- Go 1.20+
- [go-winres](https://github.com/tc-hib/go-winres)

### Development

You have to generate the icon resource file before run for the first time.

```bash
cd winres
go-winres simply --icon icon.ico --arch amd64,386,arm64
```

To build the executable, run batch file.

```bash
./build.bat
```
