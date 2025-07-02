# Pre-compiled release binary
FROM alpine:latest AS builder-release

RUN apk add --no-cache curl

ARG VERSION
ARG TARGETPLATFORM

RUN \
  case ${TARGETPLATFORM} in \
  "linux/amd64") ARCH="x86_64" ;; \
  "linux/arm64") ARCH="arm64" ;; \
  *) echo "Unsupported platform for release build: ${TARGETPLATFORM}" >&2; exit 1 ;; \
  esac; \
  wget https://github.com/5rahim/seanime/releases/download/${VERSION}/seanime-${VERSION#v}_Linux_${ARCH}.tar.gz && \
  tar -xzf seanime-*.tar.gz && \
  mv seanime seanime-server

# Binary from the latest source code
FROM golang:1.24-alpine AS builder-source

RUN apk add --no-cache git npm

WORKDIR /src
RUN git clone --depth 1 https://github.com/5rahim/seanime.git .
WORKDIR /src/seanime-web

RUN npm install
RUN npm run build

WORKDIR /src
RUN mkdir -p web && mv seanime-web/out/* web/

RUN go mod download
RUN go build -o seanime-server -trimpath -ldflags="-s -w"

# Development builds from source
FROM alpine:latest AS production-source

EXPOSE 43211

RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder-source /src/seanime-server /app/seanime

RUN chown -R app:app /app
USER app

CMD ["./seanime"]

# Release builds using pre-compiled binaries
FROM alpine:latest AS production-release

EXPOSE 43211

RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder-release /seanime-server /app/seanime

RUN chown -R app:app /app
USER app

CMD ["./seanime"]