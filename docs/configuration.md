# Configuration

## Current Status

Runtime configuration files are intentionally **not committed** to the repository. The eventual server will generate its configuration on first run in the appropriate OS-specific configuration directory.

## Intended Configuration Model

- configuration generated at runtime
- CLI flags override environment variables
- environment variables override config file values
- config file values override embedded defaults

## Planned Areas

- server networking and listen addresses
- data, cache, log, backup, and PID paths
- tenant and auth settings
- telephony and backend integration settings
- capability-driven feature exposure
- SMTP and delivery settings
