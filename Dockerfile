# SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
#
# SPDX-License-Identifier: AGPL-3.0-only

# syntax=docker/dockerfile:1
FROM gcr.io/distroless/static-debian12

ARG appName=beacon

COPY ./__build/${appName} /usr/local/bin/${appName}

ENTRYPOINT ["beacon"]
