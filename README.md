# YACS — Yet Another Cloud Server

YACS is a simple installation tool for setting up your own personal cloud server using [ownCloud Infinite Scale (oCIS)](https://owncloud.com/infinite-scale/).

It uses Docker to handle the installation and configuration, with optional [ONLYOFFICE](https://www.onlyoffice.com/) integration for document editing.

> **Development status:** Under active development. Currently supports Linux.

## Features

- Interactive installation wizard.
- Automatic Docker and system requirement checks.
- Automatic configuration and TLS certificate generation.
- Optional ONLYOFFICE integration.
- Restore existing installations without losing configuration or data.
- Uninstall while preserving data, or remove the installation completely.

## Requirements

Before using YACS, make sure you have:

- Linux
- Docker Engine
- Docker Compose v2
- OpenSSL

Go is only required if building YACS from source.

## Installation

YACS can be installed using a prebuilt binary or built from source.

### Download a release binary

Prebuilt Linux binaries are available from the [YACS GitHub Releases](https://github.com/caviarcbr/YACS/releases/latest) page.

1. Open the latest release.
2. Download the archive matching your system architecture:
   - `linux_amd64` — most Intel and AMD 64-bit systems.
   - `linux_arm64` — 64-bit ARM systems.
3. Download `checksums.txt` from the same release if you want to verify the archive.

Extract the downloaded archive:

```bash
tar -xzf yacs_VERSION_linux_ARCH.tar.gz
```

Replace `VERSION` and `ARCH` with the version and architecture in the downloaded filename.

Optionally, verify the archive checksum from the directory containing `checksums.txt`:

```bash
sha256sum -c <(grep 'yacs_VERSION_linux_ARCH.tar.gz$' checksums.txt)
```

Install the executable:

```bash
mkdir -p ~/.local/bin
install -Dm755 yacs ~/.local/bin/yacs
```

Make sure `~/.local/bin` is included in your `PATH`. You can then launch YACS with:

```bash
yacs
```

### Build from source

Clone the repository and build YACS:

```bash
git clone https://github.com/caviarcbr/YACS.git
cd YACS
go build -o yacs ./cmd/yacs
```

Run the executable:

```bash
./yacs
```

## Usage

### Interactive TUI

Running YACS without arguments launches the interactive terminal interface:

```bash
yacs
```

The main menu provides three options:

- **Install** — Start a new installation and choose whether to include ONLYOFFICE.
- **Restore** — Restore an existing installation using its saved configuration, certificates, credentials, and Docker volumes.
- **Uninstall** — Remove the running services while preserving their configuration and data, or remove the installation completely.

Follow the on-screen prompts to navigate the available options.

### CLI commands

YACS also supports command-line operations without launching the interactive interface.

#### Install

Start the installation:

```bash
yacs install
```

Follow the prompts to configure your server and choose whether to include ONLYOFFICE.

YACS will generate the required configuration, create TLS certificates, and deploy the services.

If an existing installation is detected, YACS can restore it using its existing configuration and data.

#### Uninstall

Remove the running services while keeping your configuration and data:

```bash
yacs uninstall
```

You can restore the installation later by running `yacs install` again, or selecting **Restore** from the interactive menu.

To remove the installation and its associated data:

```bash
yacs uninstall --remove-data
```

**Warning:** This permanently deletes the YACS-managed installation data. Back up anything you want to keep before proceeding.

#### Other commands

| Command | Description |
|---|---|
| `yacs` | Launch the interactive interface |
| `yacs install` | Install or resume an installation |
| `yacs uninstall` | Uninstall while preserving data |
| `yacs uninstall --remove-data` | Uninstall and remove installation data |
| `yacs version` | Display version information |
| `yacs help` | Display available commands |

## Data and configuration

YACS stores its configuration, certificates, and deployment files in:

```text
~/.local/share/yacs/
```

If `XDG_DATA_HOME` is configured, YACS uses that location instead.

Docker volumes store persistent application data.

Keep the YACS data directory if you want to restore your installation later.

## Security

- Keep your configuration directory private. It contains administrator credentials and generated secrets.
- TLS certificates may be self-signed, which can cause browser security warnings.
- Docker access grants significant privileges on Linux systems.
- Configure your firewall and network access carefully before exposing your server to the internet.

YACS does not currently configure public DNS, reverse proxies, or external network access.

## Contributing

Contributions, bug reports, and suggestions are welcome!
