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
  # Give the binary a consistent name
  mv seanime seanime-server

# Build from source code for development
FROM golang:1.22-alpine AS builder-source

RUN apk add --no-cache git npm

# This step is no longer necessary as we will use npm directly.
#   bun should have better performance, but npm is more widely supported.
# RUN npm install -g bun

WORKDIR /src
RUN git clone --depth 1 https://github.com/5rahim/seanime.git .
WORKDIR /src/seanime-web

# RUN bun install
RUN npm install
# RUN bun run build
RUN npm run build

WORKDIR /src
RUN mkdir -p web && mv seanime-web/out/* web/

RUN go mod download
RUN CGO_ENABLED=0 go build -o seanime-server -trimpath -ldflags="-s -w" cmd/seanime/main.go

# Create the minimal production image
FROM alpine:latest AS production

RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder-source /src/seanime-server /app/seanime
COPY --from=builder-source /src/web /app/web

RUN chown -R app:app /app
USER app

EXPOSE 43211

CMD ["./seanime"]