# syntax=docker/dockerfile:1
FROM gcr.io/distroless/static-debian12

ARG appName=beacon

COPY ./__build/${appName} /usr/local/bin/${appName}

ENTRYPOINT ["beacon"]
