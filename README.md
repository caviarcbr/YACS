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

### Install

Start the interactive installation:

```bash
yacs install
```

Follow the prompts to configure your server and choose whether to include ONLYOFFICE.

YACS will generate the required configuration, create TLS certificates, and deploy the services.

If an existing installation is detected, YACS can restore it using its existing configuration and data.

### Uninstall

Remove the running services while keeping your configuration and data:

```bash
yacs uninstall
```

You can restore the installation later by running `yacs install` again.

To remove the installation and its associated data:

```bash
yacs uninstall --remove-data
```

**Warning:** This permanently deletes the YACS-managed installation data. Back up anything you want to keep before proceeding.

### Other commands

| Command | Description |
|---|---|
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
