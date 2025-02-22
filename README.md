# http-proxy-syscalls

This project showcases an HTTP proxy built using Go's [syscall](https://pkg.go.dev/syscall) package.

There are 2 actors involved besides client: **proxy** + **server**. Requests made to proxy are forwarded to server.

The proxy replies to the client with the response gor from server `hello world!`.

The proxy runs on `localhost:8000`.
Server runs on `localhost:9000`.

## Testing

Make sure you're running both proxy and server.
Doing a curl `locahost:8000` should give us the response from `locahost:9000`.

```shell
➜  ~ curl localhost:9000
hello world!
```

## Why don't you use net.http?

The `https://pkg.go.dev/net/http` library is ignored on purpose, so I can learn more about network programming using
internet sockets via raw system calls.