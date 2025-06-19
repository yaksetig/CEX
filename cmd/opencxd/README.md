# opencxd

**opencxd** is the OpenCX Daemon. It runs a cryptocurrency exchange with various configurable features.
**opencxd** is closest to a "normal" centralized cryptocurrency exchange.

### Secure password usage

The daemon now supports loading the key password from an environment variable
or from standard input. Use `--keypassenv` to specify an environment variable
containing the password or `--keypasspipe` to read the password from a pipe.
The original `--keypass` option continues to work for existing configurations.
