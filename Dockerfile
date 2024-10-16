FROM gcr.io/distroless/static-debian12

COPY ./__build/indieauth-server /usr/local/bin/indieauth-server

ENTRYPOINT ["indieauth-server"]
