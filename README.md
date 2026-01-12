# Hytale Downloader

A command-line tool to download Hytale server and asset files with OAuth2 authentication.

> This is a clean-room reimplementation based on reverse engineering the official binary.

## Features

- OAuth2 device code flow authentication
- Automatic token refresh and credential caching
- Progress bar during downloads
- SHA256 checksum verification
- Multiple patchline support (release, pre-release)
- Automatic update checking

## Prerequisites

- A valid Hytale account with appropriate access

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/decom-project/hytale-downloader/releases) page.

### Build from Source

```bash
go build -ldflags="-X hytale-downloader/internal/buildinfo.Version=$(date +%Y.%m.%d)-$(git rev-parse --short HEAD) -X hytale-downloader/internal/buildinfo.Branch=release" -o hytale-downloader .
```

## Usage

### Download the Latest Version

```bash
./hytale-downloader
```

First-time setup: You'll see a URL and authorization code — open the URL in a browser to log in. The tool will automatically detect when you've authenticated and start the download.

### Check Available Version

```bash
./hytale-downloader -print-version
```

### Download to a Specific Location

```bash
./hytale-downloader -download-path /path/to/game.zip
```

### Use a Different Patchline

```bash
./hytale-downloader -patchline pre-release
```

## Command Reference

| Command | Description |
|---------|-------------|
| `./hytale-downloader` | Download latest release |
| `./hytale-downloader -print-version` | Show game version without downloading |
| `./hytale-downloader -version` | Show hytale-downloader version |
| `./hytale-downloader -check-update` | Check for hytale-downloader updates |
| `./hytale-downloader -download-path game.zip` | Download to specific file |
| `./hytale-downloader -patchline pre-release` | Download from pre-release channel |
| `./hytale-downloader -skip-update-check` | Skip the automatic update check |
| `./hytale-downloader -credentials-path /path/to/creds.json` | Use custom credentials file |

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Authentication error | Delete `.hytale-downloader-credentials.json` and re-run |
| Device code expired | Restart the tool to get a new authorization code |
| Checksum mismatch | Retry the download |
| 401 Unauthorized | Re-authenticate (delete credentials file) |
| 404 Not Found | Check patchline name & your access permissions |

## Important Notes

- **Credentials are saved** — you only need to log in once
- **Files auto-validate** — SHA256 checksum verification happens automatically
- **Keep credentials secure** — add `.hytale-downloader-credentials.json` to `.gitignore`

## License

This project is provided for educational and research purposes. Use responsibly and in accordance with Hytale's terms of service.
