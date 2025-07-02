# Builds the binary from the latest source code
FROM golang:1.24-alpine AS builder-source

RUN apk add --no-cache git npm

WORKDIR /src
RUN git clone --depth 1 https://github.com/Ju1-js/seanime.git .
WORKDIR /src/seanime-web
RUN npm install && npm run build
WORKDIR /src
RUN mkdir -p web && mv seanime-web/out/* web/
RUN go mod download && go build -o seanime-server -trimpath -ldflags="-s -w"

# Fetches a pre-compiled release binary
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
  wget https://github.com/Ju1-js/seanime/releases/download/${VERSION}/seanime-${VERSION#v}_Linux_${ARCH}.tar.gz && \
  tar -xzf seanime-*.tar.gz && \
  mv seanime seanime-server


# --- Set 1: Built from SOURCE ---
FROM alpine:latest AS production-cpu-from-source
RUN apk add --no-cache ffmpeg
EXPOSE 43211
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder-source /src/seanime-server /app/seanime
RUN chown -R app:app /app
USER app
CMD ["./seanime"]


# --- Set 2: Built from RELEASE ---
FROM alpine:latest AS production-cpu-from-release
RUN apk add --no-cache ffmpeg
EXPOSE 43211
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder-release /seanime-server /app/seanime
RUN chown -R app:app /app
USER app
CMD ["./seanime"]


FROM nvidia/cuda:12.1.1-base-ubuntu22.04 AS production-nvidia-from-release
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg && rm -rf /var/lib/apt/lists/*
EXPOSE 43211
RUN addgroup --system app && adduser --system --ingroup app --no-create-home app
WORKDIR /app
COPY --from=builder-release /seanime-server /app/seanime
RUN chown -R app:app /app
USER app
CMD ["./seanime"]