# go-qr-detector

[![go](https://github.com/sters/go-qr-detector/workflows/Go/badge.svg)](https://github.com/sters/go-qr-detector/actions?query=workflow%3AGo)
[![codecov](https://codecov.io/gh/sters/go-qr-detector/branch/main/graph/badge.svg)](https://codecov.io/gh/sters/go-qr-detector)
[![go-report](https://goreportcard.com/badge/github.com/sters/go-qr-detector)](https://goreportcard.com/report/github.com/sters/go-qr-detector)

## Usage

```
go run main.go -f {FILE PATH}
```

- Support multiple QR codes in 1 image file.
- Support various image types (png/jpeg/gif/webp)
- If QR code founds, create another file such as `{FILE PATH}_detected.png`.
