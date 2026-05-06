# Development Guide

## Current Repository Stage

This project is in the specification, bootstrap, and architectural foundation stage.

## Prerequisites

- Docker
- Make

## Build

```bash
make dev
make local
```

## Test

```bash
make test
./tests/run_tests.sh
```

## Notes

- builds and tests are container-oriented
- runtime config files are generated, not committed
- documentation should stay aligned with real implemented behavior
