# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25.9

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src

# Cache dependencies before copying source to speed up rebuilds.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# APP supports both entrypoints:
# - issue2md (CLI)
# - issue2mdweb (HTTP service, default)
ARG APP=issue2mdweb
ARG TARGETOS=linux
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    test -f "./cmd/${APP}/main.go" && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/issue2md "./cmd/${APP}"

FROM gcr.io/distroless/static-debian12:nonroot AS final
WORKDIR /app
COPY --from=builder /out/issue2md /app/issue2md

# issue2mdweb defaults to :8080 (ISSUE2MD_WEB_ADDR can override).
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/issue2md"]
