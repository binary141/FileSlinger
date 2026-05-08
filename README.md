# fileSlinger

Sling files between machines over HTTP. Run a receiver on one machine, scan the QR code on another, and drop files straight into your chosen directory — no accounts, no cloud, no fuss.

## How it works

- **serve** starts a lightweight HTTP server that accepts file uploads
- **send** pushes one or more files to a running server
- A QR code is printed in the terminal so phones and other devices can upload via browser with no app required
- A shared token authenticates every upload
- An optional **relay** mode lets you receive files through a cloud relay server when the two machines aren't on the same network

## Installation

```sh
go install github.com/binary141/fileslinger@latest
```

Or build from source:

```sh
git clone git@github.com:binary141/FileSlinger.git
cd FileSlinger
go build -o fileslinger .
```

## Usage

### Receive files (direct — same network)

```sh
fileslinger serve
```

Prints a QR code and a URL like:

```
http://192.168.1.42:8080/upload?token=aB3xQ
Dir:   uploads (unlimited)
```

Scan the QR code to upload from a phone browser, or use `send` from another machine.

**Options:**

| Flag | Default | Description |
|---|---|---|
| `-p`, `--port` | `8080` | Port to listen on |
| `-d`, `--dir` | `uploads` | Directory to save received files |
| `-n`, `--max-files` | `0` (unlimited) | Stop after receiving this many files |
| `-t`, `--token` | auto-generated | Auth token |
| `--relay` | — | Relay server URL (enables relay mode) |

### Send files (direct)

```sh
fileslinger send --host 192.168.1.42 --token aB3xQ photo.jpg document.pdf
```

**Options:**

| Flag | Default | Description |
|---|---|---|
| `-H`, `--host` | required | Server host or IP |
| `-p`, `--port` | `8080` | Server port |
| `-t`, `--token` | required | Auth token shown by `serve` |

### Relay mode (different networks)

When the two machines can't reach each other directly, use a relay server. The receiver connects to the relay over WebSocket; the sender posts to the relay's HTTP endpoint.

**Receive:**

```sh
fileslinger serve --relay https://your-relay.example.com
```

**Send:**

```sh
fileslinger send --relay https://your-relay.example.com --token <token> file.zip
```

## Configuration file

Defaults can be set in `~/.config/fileSlinger/config.json` so you don't have to repeat flags:

```json
{
  "port": 9000,
  "dir": "~/Downloads/incoming",
  "max_files": 10,
  "token": "mytoken",
  "relay_url": "https://your-relay.example.com"
}
```

Command-line flags always take precedence over the config file.

## Notes

- Files are saved with duplicate-safe naming: `photo.jpg`, `photo (1).jpg`, `photo (2).jpg`, …
- Max upload size per request is 10 GiB
- The server auto-shuts down when `--max-files` is reached
- The token is passed via the `X-Token` header (or as a `?token=` query param for browser uploads)
