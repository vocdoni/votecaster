FROM node:20-slim AS web

ARG APP_URL=${APP_URL}
ENV BASE_URL=/app
ENV APP_URL=${APP_URL}

ARG VOCDONI_COMMUNITYRESULTSADDRESS=${VOCDONI_COMMUNITYRESULTSADDRESS}
ENV VOCDONI_COMMUNITYRESULTSADDRESS=${VOCDONI_COMMUNITYRESULTSADDRESS}
ARG MAINTENANCE=${MAINTENANCE}
ENV MAINTENANCE=${MAINTENANCE}
ARG VOCDONI_AIRSTACKAPIKEY=${VOCDONI_AIRSTACKAPIKEY}
ENV VOCDONI_AIRSTACKAPIKEY=${VOCDONI_AIRSTACKAPIKEY}
ARG VOCDONI_DEGENCHAINRPC=${VOCDONI_DEGENCHAINRPC}
ENV VOCDONI_DEGENCHAINRPC=${VOCDONI_DEGENCHAINRPC}
ARG VOCDONI_ENVIRONMENT=${VOCDONI_ENVIRONMENT}
ENV VOCDONI_ENVIRONMENT=${VOCDONI_ENVIRONMENT}
ARG VOCDONI_CHAINS=${VOCDONI_CHAINS}
ENV VOCDONI_CHAINS=${VOCDONI_CHAINS}

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app
COPY webapp /app
COPY chains_config.json /

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run -r build

FROM golang:1.23 AS builder

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
COPY --from=builder /src/images images
COPY --from=web /app/dist webapp
COPY chains_config.json /app/

ENTRYPOINT ["/app/farcastervote"]
