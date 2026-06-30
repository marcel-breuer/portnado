# IPC Protocol Design

## Transport

The control protocol uses a Unix domain socket:

```text
~/Library/Application Support/Portnado/run/portnado.sock
```

The parent directory and socket must be accessible only to the current user. The daemon rejects stale sockets it cannot prove are inactive.

## Framing

Phase 2 should implement bounded newline-delimited JSON:

- UTF-8 JSON object per line.
- Maximum frame size: 1 MiB.
- Read deadline per frame.
- Unknown fields ignored only when forward-compatible.
- Malformed frames return a protocol error and close the connection when necessary.

## Request Envelope

```json
{
  "protocolVersion": 1,
  "requestId": "uuid",
  "method": "routes.list",
  "params": {}
}
```

## Response Envelope

```json
{
  "protocolVersion": 1,
  "requestId": "uuid",
  "ok": true,
  "result": {}
}
```

Error response:

```json
{
  "protocolVersion": 1,
  "requestId": "uuid",
  "ok": false,
  "error": {
    "code": "route_not_found",
    "message": "Route was not found.",
    "details": {}
  }
}
```

## Initial Method Set

Phase 2:

- `daemon.status`
- `daemon.version`
- `daemon.shutdown`

Later phases:

- `scan.run`
- `routes.list`
- `routes.approve`
- `routes.reject`
- `routes.enable`
- `routes.disable`
- `config.validate`
- `doctor.run`
- `settings.get`
- `settings.update`

## Error Codes

Error codes are stable snake_case strings. Human text may change; automation should rely on codes.

## Security

The daemon must not expose a TCP management endpoint. IPC methods must not execute arbitrary commands supplied by clients. Setup methods must be dry-run first and require explicit confirmation for changes.
