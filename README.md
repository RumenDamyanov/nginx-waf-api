# nginx-waf-api

REST API daemon for dynamic control of nginx-waf module.

> **PLANNED PROJECT** - This project is scaffolded and ready for implementation.

## Overview

nginx-waf-api is a standalone Go daemon that provides a REST API for dynamically
managing nginx-waf IP lists without manual file editing or nginx configuration changes.

## Features (Planned)

- REST API for IP list management (CRUD operations)
- API key authentication
- Automatic nginx reload after changes
- Audit logging of all modifications
- Prometheus metrics endpoint
- Atomic file operations (no partial writes)

## Architecture

```
API Client ──> nginx-waf-api ──> IP List Files ──> nginx-waf
 (UI/CLI)     (Go daemon)     (/etc/nginx/...)    (C module)
```

## API Endpoints (Planned)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/lists` | List all IP lists |
| GET | `/api/v1/lists/{name}` | Get list details |
| POST | `/api/v1/lists/{name}/entries` | Add IP to list |
| DELETE | `/api/v1/lists/{name}/entries/{ip}` | Remove IP from list |
| POST | `/api/v1/reload` | Trigger nginx reload |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Quick Start

```bash
# Build
make build

# Configure
cp conf/config.example.yaml /etc/nginx-waf-api/config.yaml
# Edit configuration...

# Run
./nginx-waf-api -config /etc/nginx-waf-api/config.yaml
```

## Installation

### From OBS packages (recommended)

Available for Fedora, openSUSE, Debian, and Ubuntu via
[OBS](https://build.opensuse.org/package/show/home:rumenx/nginx-waf-api).

### From source

```bash
make build
sudo make install
sudo cp conf/config.example.yaml /etc/nginx-waf-api/config.yaml
sudo cp dist/nginx-waf-api.service /etc/systemd/system/
sudo systemctl enable --now nginx-waf-api
```

## Related Projects

- [nginx-waf](https://github.com/RumenDamyanov/nginx-waf) - Core nginx module (required)
- [nginx-waf-feeds](https://github.com/RumenDamyanov/nginx-waf-feeds) - Threat feed updater
- [nginx-waf-lua](https://github.com/RumenDamyanov/nginx-waf-lua) - OpenResty Lua integration
- [nginx-waf-ui](https://github.com/RumenDamyanov/nginx-waf-ui) - Web management interface

## License

BSD 3-Clause License - see [LICENSE.md](LICENSE.md) for details.
