FROM golang:alpine AS builder

RUN apk add --no-cache git bash

ARG TARGETARCH
ARG VERSION=dev
ARG BUILD_DATE
ARG COMMIT_ID
ARG OFFICIALSITE=""

WORKDIR /build
RUN mkdir -p /build/binary

COPY go.mod ./
RUN go mod download

COPY src/ ./src/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags "-s -w -X 'main.Version=${VERSION}' -X 'main.CommitID=${COMMIT_ID}' -X 'main.BuildDate=${BUILD_DATE}' -X 'main.OfficialSite=${OFFICIALSITE}'" \
    -o /build/binary/caspbx ./src

FROM alpine:latest

ARG VERSION=dev
ARG BUILD_DATE
ARG COMMIT_ID
ARG LICENSE=MIT
ARG MAINTAINER_NAME
ARG MAINTAINER_EMAIL
ARG PLATFORM_REPO_URL

LABEL maintainer="${MAINTAINER_NAME} <${MAINTAINER_EMAIL}>" \
      org.opencontainers.image.vendor="casapps" \
      org.opencontainers.image.authors="casapps" \
      org.opencontainers.image.title="caspbx" \
      org.opencontainers.image.base.name="caspbx" \
      org.opencontainers.image.description="caspbx - standard image (alpine)" \
      org.opencontainers.image.licenses="${LICENSE}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.schema-version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT_ID}" \
      org.opencontainers.image.url="${PLATFORM_REPO_URL}" \
      org.opencontainers.image.source="${PLATFORM_REPO_URL}" \
      org.opencontainers.image.documentation="${PLATFORM_REPO_URL}" \
      org.opencontainers.image.vcs-type="Git" \
      com.github.containers.toolbox="false"

RUN apk add --no-cache \
    git \
    curl \
    bash \
    tini \
    tor

COPY --from=builder /build/binary/caspbx /usr/local/bin/caspbx
COPY docker/file_system/ /
COPY docker/Dockerfile /root/Dockerfile

RUN chmod 755 /usr/local/bin/*

ENV MODE=development \
    TZ=America/New_York \
    DATABASE_DIR=/data/db/sqlite

EXPOSE 80

STOPSIGNAL SIGRTMIN+3

HEALTHCHECK --start-period=10m --interval=5m --timeout=15s --retries=3 \
    CMD /usr/local/bin/caspbx --status || exit 1

ENTRYPOINT [ "tini", "-p", "SIGTERM", "--", "/usr/local/bin/entrypoint.sh" ]
