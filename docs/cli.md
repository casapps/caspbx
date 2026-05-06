# CLI Reference

## Current Status

The repository currently contains only the bootstrap `caspbx` server scaffold. The companion `caspbx-cli` and `caspbx-agent` binaries are required by the spec but are not implemented yet.

## Planned Binaries

- `caspbx` — server binary
- `caspbx-cli` — user/admin client
- `caspbx-agent` — optional remote/collector agent

## Current Bootstrap Flags

The bootstrap server scaffold currently accepts the high-level flag surface needed for early build and test work, including:

- `--help`
- `--version`
- `--status`
- path, mode, output, and service-related flag placeholders
