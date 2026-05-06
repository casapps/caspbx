# Installation

## Current Status

The production runtime is not implemented yet. This repository currently provides the bootstrap scaffolding for the eventual server, client, documentation, and test flows.

## Planned Installation Targets

### Docker

The project spec requires production Docker assets under `docker/`.

### Native Binary

The project spec requires a single static `caspbx` server binary built from `./src`.

### Service Installation

The project spec requires service-management support for native deployments after the runtime server is implemented.

## Development Bootstrap

For now, the repository can be bootstrapped and verified with:

```bash
make local
make test
./tests/run_tests.sh
```
