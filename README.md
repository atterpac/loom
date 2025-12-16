# Loom

> [!CAUTION]
> This is a personal tool and not fully tested. Use with caution.

A terminal UI client for [Temporal](https://temporal.io).

## Installation

```bash
go install github.com/atterpac/loom/cmd/loom@latest
```

## Usage

```bash
loom --address localhost:7233
```

### Flags

| Flag | Description |
|------|-------------|
| `--address` | Temporal server address (host:port) |
| `--namespace` | Default namespace |
| `--profile` | Connection profile name |
| `--tls-cert` | Path to TLS certificate |
| `--tls-key` | Path to TLS private key |
| `--tls-ca` | Path to CA certificate |
| `--theme` | Theme name |

### Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `Enter` | Select |
| `Esc` | Go back |
| `T` | Theme selector |
| `P` | Profile selector |
| `?` | Help |
| `q` | Quit |

## Themes

26 built-in themes including TokyoNight, Catppuccin, Dracula, Nord, Gruvbox, and more.
