# httpserver

A http server for testing purposes.

## Install

```console
$ go install github.com/masa213f/httpserver@latest
```

## Usage

Run as a server that returns a fixed string.

```console
$ httpserver -mode text="hello world"
```

Run as a file server. (by using [`http.FileServer`](https://pkg.go.dev/net/http#FileServer))

```console
$ httpserver -mode file=/path/to/public_html
```

Run as a server that dumps a http request. (by using [`httputil.DumpRequest`](https://pkg.go.dev/net/http/httputil#DumpRequest))

```console
$ httpserver -mode dump
```

Listen on a specified port.

```console
$ httpserver -listen :2026
```

Listen on a unix domain socket.

```console
$ httpserver -listen unix:/path/to/socket
```
