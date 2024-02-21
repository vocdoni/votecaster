FROM node:16 AS web

ARG APP_URL=${APP_URL}
ENV BASE_URL=/app
ENV APP_URL=${APP_URL}

WORKDIR /app
COPY webapp /app

RUN npm install -f
RUN npm run build

FROM golang:1.22 AS builder

WORKDIR /src
ENV CGO_ENABLED=1
RUN go env -w GOCACHE=/go-cache
COPY . .
RUN --mount=type=cache,target=/go-cache go mod download
RUN --mount=type=cache,target=/go-cache go build -o=farcastervote -ldflags="-s -w"

FROM debian:bookworm-slim as base

WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Support for go-rapidsnark witness calculator (https://github.com/iden3/go-rapidsnark/tree/main/witness)
COPY --from=builder /go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so \
                    /go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so

# Support for go-rapidsnark prover (https://github.com/iden3/go-rapidsnark/tree/main/prover)
RUN apt-get update && \
	apt-get install --no-install-recommends -y libc6-dev libomp-dev openmpi-common libgomp1 curl && \
	apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /src/farcastervote ./
COPY --from=builder /src/farcaster_census.json ./
COPY --from=builder /src/fonts fonts
COPY --from=builder /src/images images
COPY --from=web /app/dist webapp

ENTRYPOINT ["/app/farcastervote"]
