<!--
SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>

SPDX-License-Identifier: CC-BY-4.0
-->

# Project Beacon

## Overview

Beacon is a standalone [IndieAuth](https://indieauth.net/) provider written in [Go](https://go.dev/).

This project is currently in early stages of development so there may be breaking changes as the project evolves.

## Development

Beacon is actively developed in [Code Flow](https://codeflow.dananglin.me.uk/apollo/beacon) with the `main` branch synced to the following forges:

- [**Codeberg**](https://codeberg.org/dananglin/beacon)
- [**GitHub**](https://github.com/dananglin/beacon)

## Full documentation

The project's documentation can be found [here](https://dananglin.me.uk/projects/beacon/).

## 🤝 Contributing

### Run the application

Clone the repository.

```bash
git clone https://github.com/dananglin/beacon.git
cd beacon
```

Copy the example configuration and generate the JWT secret.

```bash
cp ./example/config.json ./config.json
openssl rand -base64 36
```

Update the configuration with the secret and run the application.

```bash
./beacon serve --config config.json
```

### Run the tests

You can run the tests with the below command.

```bash
go test -v ./...
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.

## Licensing

This project is REUSE compliant so the copyright and licensing information is stored in the header in every file,
but in general:

- All original source code is licensed under `AGPL-3.0-only`.
- All documentation is licensed under `CC-BY-4.0`.
- Configuration and data files are licensed under `CC0-1.0`.
