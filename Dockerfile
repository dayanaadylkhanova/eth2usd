# --- build ---
FROM golang:1.24.7-alpine AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod tidy && CGO_ENABLED=0 go build -o /out/eth2usd ./cmd/eth2usd

# --- runtime ---
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/eth2usd /usr/local/bin/eth2usd
ENTRYPOINT ["/usr/local/bin/eth2usd"]
