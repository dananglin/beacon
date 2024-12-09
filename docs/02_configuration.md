+++
title = "Configuration reference"
description = "The configuration reference documentation."
weight = 2
slug = "configuration-reference"
template = "project-page.html"
+++
<!--
SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>

SPDX-License-Identifier: CC-BY-4.0
-->

# Configuration reference

## Config

| Field | Type | Description|
|-------|------|------------|
| `bindAddress` | string | The IP address to bind your Beacon instance to. |
| `port` | int | The port to listen for HTTP requests. |
| `domain` | string | The resolvable domain name of your Beacon instance. |
| `database` | [DatabaseConfig](#databaseconfig) | Database configuration. |
| `jwt` | [JWTConfig](#jwtconfig) | JWT configuration. |

## DatabaseConfig

| Field | Type | Description|
|-------|------|------------|
| `path` | string | The (absolute) path to the database file. |

## JWTConfig

| Field | Type | Description|
|-------|------|------------|
| `secret` | string | The randomly generated JWT secret key for signing JWTs.<br>You can generate a key for your instance using this example command:<br>`openssl rand -base64 32`|
| `cookieName` | string | The name of the cookie to store your JWT session token. You can use any alpha-numeric characters plus `+`, `-`, `.`, and `_`.<br>By default the cookie name is set to `beacon_is_great`.|
