# ocx

**ocx** is a command-line client for many RPC commands which OpenCX RPC packages support.
**ocx** is currently compatible with both commands in `cxrpc` as well as some in `cxauctionrpc`, so it can be used for both servers running `frred` or `opencxd`.

### Secure password usage

To unlock your client key without exposing the password on the command line you
can provide the password via an environment variable or pipe. Use `--keypassenv`
to specify the environment variable or `--keypasspipe` to read from standard
input. The legacy `--keypass` option is still supported for compatibility.
