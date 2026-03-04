<!--
SPDX-FileCopyrightText: 2025 Dan Anglin <d.n.i.anglin@gmail.com>

SPDX-License-Identifier: CC-BY-4.0
-->

![A screenshot of Beacon](assets/project_screenshot.png "Project Beacon")

Beacon is a standalone [IndieAuth](https://indieauth.net/) provider which closely follows the latest version
of the [IndieAuth specification](https://indieauth.spec.indieweb.org/).

With Beacon you can use your own domain name to sign into IndieAuth supported applications or websites such
as [Owncast](https://owncast.online/).
Beacon would be suitable for you if your domain or website does not have an IndieAuth provider built in.

Setting up your own Beacon instance is relatively easy.
A typical setup process would involve:

- Building the binary or Docker image.
- Deploying the application behind a reverse proxy.
- Setting up your profile.
- Updating your domain's HTTP server or website's HTML page by adding a link to your instance's authorization and token endpoints.

A detailed guide for the setup process can be found [here](setup.md).

## Features

- Single user deployment
- Authorization endpoint
- Token endpoint

## Development

This project is actively developed on my [**Code Flow**](https://codeflow.dananglin.me.uk/apollo/beacon) instance
with the `main` branch mirrored to [**Codeberg**](https://codeberg.org/dananglin/beacon).

## Licensing

This project is REUSE compliant so the copyright and licensing information is stored in the header in every file.
In general:

- All original source code is licensed under `AGPL-3.0-only`.
- All documentation is licensed under `CC-BY-4.0`.
- Configuration and data files are licensed under `CC0-1.0`.
