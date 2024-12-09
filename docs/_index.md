+++
title = "Beacon"
description = "A standalone IndieAuth provider written in Go."
weight = 0
sort_by = "weight"
template = "section-project.html"
+++

## Summary

Beacon is an [IndieAuth](https://indieauth.net/) provider which closely follows the latest version of the [IndieAuth specification](https://indieauth.spec.indieweb.org/).

With Beacon you can use your own domain name to sign into IndieAuth supported applications or websites such as [Owncast](https://owncast.online/).
Beacon would be suitable for you if your domain or website does not have IndieAuth built in.

Setting up your own Beacon instance is relatively easy:

- Build the binary or Docker image.
- Deploy it behind your reverse proxy.
- Set up your profile.
- Update your domain's HTTP server or website's HTML page by adding a link to your instance's authorization and token endpoints.
- Sign into an IndieAuth supported application or website.

A guide to the setup process can be found [here](@/projects/beacon/01_setup_guide.md).

### Features

- Single user deployment
- Authorization endpoint
- Token endpoint

## Development

This project is actively developed in [Code Flow](https://codeflow.dananglin.me.uk/apollo/beacon) with the
`main` branch synced to the following forges:

- [**Codeberg**](https://codeberg.org/dananglin/beacon)
- [**GitHub**](https://github.com/dananglin/beacon)

## Licensing

This project is REUSE compliant so the copyright and licensing information is stored in the header in every file. In general:

- All original source code is licensed under `AGPL-3.0-only`.
- All documentation is licensed under `CC-BY-4.0`.
- Configuration and data files are licensed under `CC0-1.0`.
