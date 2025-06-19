# webui

**webui** provides a very small web interface for OpenCX. It exposes a few RPC
methods over HTTP and serves a basic HTML page to display the orderbook.

Run it with:

```sh
go build ./cmd/webui/...
./webui
```

The server listens on port `8080` by default and assumes the exchange RPC server
is reachable on `localhost:12345`.
