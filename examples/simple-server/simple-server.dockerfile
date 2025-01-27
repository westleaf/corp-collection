FROM golang:1.23 as BUILDER

WORKDIR /app

RUN mkdir -p /build

COPY . ./

RUN CDO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -o main .

FROM scratch

COPY --from=BUILDER /app/main /app/main

LABEL org.opencontainers.image.source="https://github.com/westleaf/corp-collection"
LABEL org.opencontainers.image.title="simple-unsecure-fileserver"
LABEL org.opencontainers.image.description="A simple but unsecure fileserver"

ENTRYPOINT ["/app/main"]
