# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e
FROM golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# No implicit toolchain downloads and no on-the-fly go.mod/go.sum edits.
ENV GOTOOLCHAIN=local \
    GOFLAGS=-mod=readonly

WORKDIR /src
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false \
    -ldflags="-s -w \
      -X github.com/Kandru/cs2-demo-analyzer/internal/version.Version=${VERSION} \
      -X github.com/Kandru/cs2-demo-analyzer/internal/version.Commit=${COMMIT} \
      -X github.com/Kandru/cs2-demo-analyzer/internal/version.Date=${BUILD_DATE}" \
    -o /out/cs2-demo-analyzer ./cmd/cs2-demo-analyzer

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 1000 app
COPY --from=build /out/cs2-demo-analyzer /usr/local/bin/cs2-demo-analyzer
# The image runs as root by default so bind-mounted output directories work on
# any host. Run with `--user $(id -u):$(id -g)` if you prefer a non-root UID.
WORKDIR /reports
ENTRYPOINT ["cs2-demo-analyzer"]
CMD ["watch", "--input", "/demos", "--output", "/out"]
