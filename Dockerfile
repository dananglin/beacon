FROM gcr.io/distroless/static-debian12

COPY ./__build/indieauth-server /indieauth-server

ENTRYPOINT ["/indieauth-server"]
