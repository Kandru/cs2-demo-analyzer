# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /src
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false \
    -ldflags="-s -w \
      -X github.com/Kandru/cs2-demo-analyzer/internal/version.Version=${VERSION} \
      -X github.com/Kandru/cs2-demo-analyzer/internal/version.Commit=${COMMIT} \
      -X github.com/Kandru/cs2-demo-analyzer/internal/version.Date=${BUILD_DATE}" \
    -o /out/cs2-battlelog ./cmd/cs2-battlelog

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 1000 app
COPY --from=build /out/cs2-battlelog /usr/local/bin/cs2-battlelog
# The image runs as root by default so bind-mounted output directories work on
# any host. Run with `--user $(id -u):$(id -g)` if you prefer a non-root UID.
WORKDIR /reports
ENTRYPOINT ["cs2-battlelog"]
CMD ["watch", "--input", "/demos", "--output", "/out"]
